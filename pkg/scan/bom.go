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
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/source"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
	"gitlab.bit9.local/octarine/cbctl/pkg/model/bom"
)

const (
	digestToTag   = ":sha256_"
	digestStart   = "@sha256:"
	tarFileEnding = ".tar"
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
	if opts.BypassDockerDaemon {
		if tempDir, creationErr := createTempDir(); creationErr == nil {
			defer func() {
				if rmErr := os.RemoveAll(tempDir); rmErr != nil {
					logrus.Errorf("Failed to remove cached directory: %v", rmErr)
				}
			}()

			// replace the input with image we copied
			if cachedImageDir, e := h.copyImage(originalInput, tempDir, opts); e == nil {
				logrus.WithFields(logrus.Fields{"original-input": originalInput, "new-input": cachedImageDir}).
					Info("replace origin input after copy with cached image path")

				input = cachedImageDir
			}
		}
	}

	if inputSrc == image.OciDirectorySource || inputSrc == image.UnknownSource {
		e := cberr.NewError(cberr.SBOMGenerationErr, "The input is not a valid scan source", err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	return GenerateSBOMFromInput(input, originalInput, opts.FullTag)
}

// GenerateSBOMFromInput create bom struct after coping.
func GenerateSBOMFromInput(input, originalInput, forceFullTag string) (*Bom, error) {
	var exclusions []string

	theSource, cleanup, err := source.New(input, &image.RegistryOptions{}, exclusions)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to generate source from input %v", input)

		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	// clean up sterescope tmp files if needed
	defer cleanup()

	cft := cataloger.DefaultConfig()

	theCatalog, _, linuxDistro, err := syft.CatalogPackages(theSource, cft)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to catalog input %v", input)

		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	doc, err := bom.NewJSONDocument(theCatalog, theSource.Metadata, linuxDistro, source.SquashedScope)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse the sbom for %v", input)
		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	// fullTag is a tag that use for image-scanning-pipeline (checking ShouldIgnore/ForceScan and updating scan status).
	// target.Tags is sent to anchore and will determine which tags the webhook will update.
	var fullTag string

	target, ok := doc.Source.Target.(bom.JSONImageSource)
	if !ok {
		return nil, fmt.Errorf("failed to convert taget to image metadata type")
	}

	target.Tags = formatTags(target.Tags, forceFullTag)
	if len(target.Tags) > 0 {
		logrus.WithField("tags", target.Tags).Debug("Valid tags detected from sbom")
		// pick last so if there is force tag it will be used
		fullTag = revertAnchoreDigestChange(target.Tags[len(target.Tags)-1])
	} else {
		// attach tag to bom because not exists.
		fullTag = generateFullTagFromOriginInput(originalInput, target.ManifestDigest)
		generateTag, formattingErr := formatTag(fullTag)
		if formattingErr != nil {
			logrus.WithFields(logrus.Fields{"err": formattingErr, "originalInput": originalInput}).
				Error("fail formatting originalInput into tag")
			return nil, formattingErr
		}
		target.Tags = []string{generateTag}
	}

	doc.Source.Target = target

	logrus.WithField("fullTag", fullTag).Infof("SBOM generated successfully")

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

func addDefaultValuesToFullTag(tag string) (string, error) {
	ref, err := reference.ParseDockerRef(tag)
	if err != nil {
		return tag, err
	}

	return ref.String(), err
}

// formatTag try to format the tags that Syft stores at SBOM.
// adding default tag plus repo to tags or editing digested tag.
// anchore cant handle tags with @ (they need all images to be tag and not digested).
func formatTag(tag string) (string, error) {
	if strings.Contains(tag, digestStart) {
		tag = strings.ReplaceAll(tag, digestStart, digestToTag)
	}

	return addDefaultValuesToFullTag(tag)
}

// formatTags format the attach tags and add opts fullTag if exists.
func formatTags(tags []string, forceFullTag string) []string {
	if forceFullTag != "" {
		// the FullTag need to be last
		tags = append(tags, forceFullTag)
	}

	formattedTags := make([]string, 0, len(tags))

	for _, tag := range tags {
		formattedTag, err := formatTag(tag)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err, "tag": tag}).
				Warning("fail formatting tag (remove from tags)")
			continue
		}

		formattedTags = append(formattedTags, formattedTag)
	}

	return formattedTags
}

// revertAnchoreDigestChange revert any change made to digest in order to create image scanning pipeline fullTag.
func revertAnchoreDigestChange(anchorNormalizedTag string) string {
	return strings.ReplaceAll(anchorNormalizedTag, digestToTag, digestStart)
}

func generateFullTagFromOriginInput(tag string, manifestDigest string) string {
	// since we already generated sbom with this input with sha256, the input can be guaranteed as
	// an image with valid sha256 digest, no need to further checking here
	if strings.Contains(tag, tarFileEnding) {
		repo := strings.TrimSuffix(tag[strings.LastIndex(tag, "/")+1:], tarFileEnding)
		newTag := strings.ReplaceAll(manifestDigest, "sha256:", "")
		tag = fmt.Sprintf("%s:%s", repo, newTag)
	}

	tag, err := addDefaultValuesToFullTag(tag)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err, "tag": tag}).
			Warning("fail formatting tag while generating FullTag from origin input")
	}

	return tag
}
