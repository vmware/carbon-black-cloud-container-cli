/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package scan

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/source"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

// Bom contains the full bill of materials for an image, along with some
// additional helpful metadata.
type Bom struct {
	// FullTag is the full tag of the bom
	FullTag string
	// ManifestDigest is the sha256 of this image manifest json
	ManifestDigest string
	// Packages enumerates the packages in the bill of materials
	Packages bom.JSONDocument
}

// RegistryHandler coordinates with OCI registry APIs in order to retrieve
// container images as needed.
type RegistryHandler struct {
	registryImageCopy func(
		ctx context.Context,
		policyContext *signature.PolicyContext,
		destRef,
		srcRef types.ImageReference,
		options *copy.Options,
	) (copiedManifest []byte, retErr error)
}

// NewRegistryHandler constructs a new RegistryHandler instance.
func NewRegistryHandler() RegistryHandler {
	return RegistryHandler{registryImageCopy: copy.Image}
}

// Generate leverages Syft and Stereoscope to generate a bill of materials for a given image tag.
func (h *RegistryHandler) Generate(originalInput string, opts Option) (*Bom, error) {
	eventBus := partybus.NewBus()
	subscription := eventBus.Subscribe()
	input := originalInput

	stereoscope.SetBus(eventBus)
	syft.SetBus(eventBus)

	go func() {
		for e := range subscription.Events() {
			eventType := bus.EventType(e.Type)
			bus.Publish(bus.NewEvent(eventType, e.Value, false))
		}
	}()

	inputSrc, _, err := image.DetectSource(input)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to detect the source of input %v", originalInput)
		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	// error happened in this block will be ignored, since if we failed to copy without docker daemon,
	// we will pass the input to Syft and let Syft pull image by docker daemon
	if !opts.UseDockerDaemon {
		if tempDir, creationErr := createTempDir(); creationErr == nil {
			defer func() {
				if rmErr := os.RemoveAll(tempDir); rmErr != nil {
					logrus.Errorf("Failed to remove cached directory: %v", rmErr)
				}
			}()

			// replace the input with image we copied
			if cachedImageDir, e := h.copyImage(originalInput, tempDir, opts); e == nil {
				input = cachedImageDir
			}
		}
	}

	if inputSrc == image.OciDirectorySource || inputSrc == image.UnknownSource {
		e := cberr.NewError(cberr.SBOMGenerationErr, "The input is not a valid scan source", err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	theSource, cleanup, err := source.New(input, &image.RegistryOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to generate source from input %v", originalInput)
		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	// clean up sterescope tmp files if needed
	defer cleanup()

	theCatalog, theDistro, err := syft.CatalogPackages(theSource, source.SquashedScope)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to catalog input %v", originalInput)
		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	doc, err := bom.NewJSONDocument(theCatalog, theSource.Metadata, theDistro, source.SquashedScope)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse the sbom for %v", originalInput)
		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	// attach tag to bom if not exists, use original input here for getting tag
	fullTag := attachTag(&doc, originalInput, opts)

	logrus.Infof("SBOM generated successfully with full tag: %v", fullTag)

	return &Bom{
		FullTag:        fullTag,
		ManifestDigest: theSource.Metadata.ImageMetadata.ManifestDigest,
		Packages:       doc,
	}, nil
}

func createTempDir() (string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-cache", internal.ApplicationName))
	if err != nil {
		e := cberr.NewError(cberr.SBOMGenerationErr, "could not create temp dir", err)
		return "", e
	}

	logrus.Infof("Temp dir for cache has been created: %s", dir)

	return dir, nil
}

func (h *RegistryHandler) copyImage(input, cacheDir string, opts Option) (string, error) {
	stage := &progress.Stage{Current: "initializing"}
	prog := &progress.Manual{Total: 1}
	value := progress.StagedProgressable(&struct {
		progress.Stager
		progress.Progressable
	}{
		Stager:       stage,
		Progressable: prog,
	})
	bus.Publish(bus.NewEvent(bus.CopyImage, value, false))

	defer prog.SetCompleted()

	srcRef, err := alltransports.ParseImageName("docker://" + input)
	if err != nil {
		stage.Current = "invalid source name, try with docker pull"
		e := cberr.NewError(cberr.SBOMGenerationErr, fmt.Sprintf("invalid source name %s", input), err)
		logrus.Error(e.Error())

		return "", e
	}

	srcCtx := &types.SystemContext{
		// if a multi-arch image detected, pull the linux image by default
		ArchitectureChoice:          "amd64",
		OSChoice:                    "linux",
		DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
	}

	if opts.Credential != "" {
		username, password := opts.parseAuth()
		srcCtx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	destDir := path.Join(cacheDir, "temp.tar")

	destRef, err := alltransports.ParseImageName("docker-archive:" + destDir)
	if err != nil {
		stage.Current = "invalid destination name, try with docker pull"
		e := cberr.NewError(cberr.SBOMGenerationErr, fmt.Sprintf("invalid destination name %s", input), err)
		logrus.Error(e.Error())

		return "", e
	}

	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	policyContext, _ := signature.NewPolicyContext(policy)

	copyProgress := make(chan types.ProgressProperties)

	go func() {
		for e := range copyProgress {
			percentage := math.Floor(100 * float64(e.Offset) / float64(e.Artifact.Size))
			if percentage > 100 {
				percentage = 100
			}

			msg := fmt.Sprintf("copying from %v - %v%%", e.Artifact.Digest[:18], percentage)
			stage.Current = msg
		}
	}()

	if _, err = h.registryImageCopy(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		SourceCtx:        srcCtx,
		ProgressInterval: 150 * time.Millisecond,
		Progress:         copyProgress,
	}); err != nil {
		stage.Current = "failed to copy image, try with docker pull"
		e := cberr.NewError(cberr.SBOMGenerationErr, "failed to copy image", err)
		logrus.Error(e.Error())

		return "", e
	}

	stage.Current = "copying image on disk"

	logrus.Infof("Image copied to %s successfully", destDir)

	return destDir, nil
}

// attachTag will attach a tag for those results without a tag. Return will be the normalized full-tag.
func attachTag(doc *bom.JSONDocument, input string, opts Option) (returnTag string) {
	target, ok := doc.Source.Target.(bom.JSONImageSource)
	if !ok {
		logrus.Errorf("Failed to convert taget to image metadata type")
		return input
	}

	if opts.FullTag != "" {
		target.Tags = append([]string{opts.FullTag}, target.Tags...)
		doc.Source.Target = target
	}

	generatedTag := input

	if len(target.Tags) > 0 {
		logrus.Infof("Valid tags detected from sbom: %+v", target.Tags)

		// if the tag can be normalized, we will use it as the tag
		if ref, err := reference.ParseNormalizedNamed(target.Tags[0]); err == nil {
			fullTag := reference.TagNameOnly(ref).String()
			target.Tags[0] = fullTag
			doc.Source.Target = target

			return fullTag
		}

		generatedTag = target.Tags[0]
	}

	switch {
	// since we already generated sbom with this input with sha256, the input can be guaranteed as
	// an image with valid sha256 digest, no need to further checking here
	case strings.Contains(input, "@sha256:"):
		if ref, err := reference.ParseDockerRef(input); err == nil {
			generatedTag = strings.ReplaceAll(ref.String(), "@sha256:", ":sha256_")
			returnTag = ref.String()
		}
	case strings.Contains(input, ".tar"):
		repo := strings.TrimSuffix(input[strings.LastIndex(input, "/")+1:], ".tar")
		tag := strings.ReplaceAll(target.ManifestDigest, "sha256:", "")
		generatedTag = fmt.Sprintf("%s:%s", repo, tag)
	default:
		if ref, err := reference.ParseNormalizedNamed(generatedTag); err == nil {
			generatedTag = reference.TagNameOnly(ref).String()
		}
	}

	target.Tags = []string{generatedTag}
	doc.Source.Target = target

	logrus.Infof("Created a fake tag for %v: %v", input, generatedTag)

	if returnTag == "" {
		returnTag = generatedTag
	}

	return returnTag
}
