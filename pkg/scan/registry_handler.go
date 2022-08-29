package scan

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/docker/docker/errdefs"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/syft"
	"github.com/containers/image/v5/copy"
	containersimage "github.com/containers/image/v5/image"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	imagetype "github.com/containers/image/v5/types"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

const (
	digestToTag           = ":sha256_"
	digestStart           = "@sha256:"
	tarFileEnding         = ".tar"
	tempTarFileName       = "temp.tar"
	tempLayersTarFileName = "layers_temp.tar" // TODO: Refactor
)

// RegistryHandler coordinates with OCI registry APIs in order to retrieve
// container images as needed.
type RegistryHandler struct {
	registryImageCopy func(
		ctx context.Context,
		policyContext *signature.PolicyContext,
		destRef,
		srcRef imagetype.ImageReference,
		options *copy.Options,
	) (copiedManifest []byte, retErr error)
}

// NewRegistryHandler constructs a new RegistryHandler instance.
func NewRegistryHandler() RegistryHandler {
	return RegistryHandler{
		registryImageCopy: copy.Image,
	}
}

// GenerateSBOM leverages Syft and Stereoscope to generate a bill of materials for a given image tag.
func (h *RegistryHandler) GenerateSBOM(input string, opts Option) (*Bom, error) {
	eventBus := partybus.NewBus()
	subscription := eventBus.Subscribe()
	originalInput := input

	stereoscope.SetBus(eventBus)
	syft.SetBus(eventBus)

	go func() {
		for e := range subscription.Events() {
			eventType := bus.EventType(e.Type)
			bus.Publish(bus.NewEvent(eventType, e.Value, false))
		}
	}()

	_, err := parseSourceFromRawRef(input)
	if err != nil {
		return nil, err
	}

	// error happened in this block will be ignored, since if we failed to copy without docker daemon,
	// we will pass the input to Syft and let Syft pull image by docker daemon
	if opts.BypassDockerDaemon {
		if tempDir, creationErr := createTempDir(); creationErr == nil {
			defer cleanTempDir(tempDir)

			// replace the input with image we copied
			if cachedImageDir, e := h.copyImage(originalInput, tempDir, opts); e == nil {
				logrus.WithFields(logrus.Fields{"original-input": originalInput, "new-input": cachedImageDir}).
					Info("replace origin input after copy with cached image path")

				input = cachedImageDir
			}
		}
	}

	return GenerateSBOMFromInput(input, originalInput, opts.FullTag)
}

func (h *RegistryHandler) GenerateLayers(input string, opts Option) ([]layers.Layer, error) {
	tempDir, creationErr := createTempDir()
	if creationErr != nil {
		return nil, cberr.NewError(cberr.LayersGenerationErr, "error while creating cache folder for layer generation", creationErr)
	}
	defer cleanTempDir(tempDir)
	tarPath := fmt.Sprintf("%v/%v", tempDir, tempLayersTarFileName)

	inputSrc, err := parseSourceFromRawRef(input)
	if err != nil {
		return nil, cberr.NewError(cberr.LayersGenerationErr, err.Error(), err)
	}

	img, err := copyImgToLocalTarForLayers(inputSrc, input, opts, tarPath)
	if err != nil {
		return nil, cberr.NewError(cberr.LayersGenerationErr, "Failed to copy image to local tar for layer computation", err)
	}

	manifest, _, err := img.Manifest(context.Background())
	if err != nil {
		return nil, cberr.NewError(cberr.LayersGenerationErr, "Failed to read manifest from cached image", err)
	}
	manifestDigest := fmt.Sprintf("%x", sha256.Sum256(manifest))
	config, err := img.ConfigBlob(context.Background())
	if err != nil {
		return nil, cberr.NewError(cberr.LayersGenerationErr, "Failed to read config file from cached image", err)
	}

	layers, err := GenerateLayers(manifestDigest, tarPath, config)
	if err != nil {
		err = cberr.NewError(cberr.LayersGenerationErr, "error generating image layers", err)
	}

	return layers, err
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

	srcCtx := makeDefaultSystemPullContext()

	if opts.Credential != "" {
		username, password := opts.parseAuth()
		srcCtx.DockerAuthConfig = &imagetype.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	destDir := path.Join(cacheDir, tempTarFileName)

	destRef, err := alltransports.ParseImageName("docker-archive:" + destDir)
	if err != nil {
		stage.Current = "invalid destination name, try with docker pull"
		e := cberr.NewError(cberr.SBOMGenerationErr, fmt.Sprintf("invalid destination name %s", input), err)
		logrus.Error(e.Error())

		return "", e
	}

	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	policyContext, _ := signature.NewPolicyContext(policy)

	copyProgress := make(chan imagetype.ProgressProperties)

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

func createTempDir() (string, error) {
	// TODO: This avoids multiple callers deleting the same folder in most cases but is ugly
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-cache-%d", internal.ApplicationName, rand.Int()))
	if err != nil {
		return "", fmt.Errorf("could not create temp dir: %w", err)
	}

	logrus.Infof("Temp dir for cache has been created: %s", dir)

	return dir, nil
}

func cleanTempDir(tempDir string) {
	if rmErr := os.RemoveAll(tempDir); rmErr != nil {
		logrus.Errorf("Failed to remove cached directory: %v", rmErr)
	}
}

func parseSourceFromRawRef(reference string) (image.Source, error) {
	// Try to determinate if the source is an archive
	inputSrc, err := image.DetectSourceFromPath(reference)
	if err == nil {
		if inputSrc == image.UnknownSource || inputSrc == image.OciDirectorySource {
			err = fmt.Errorf("input is not a valid archive source or image pull source")
		} else {
			return inputSrc, nil
		}
	}

	// If the source isn't an archive, try to detect the image pull source
	inputSrc = image.DetermineDefaultImagePullSource(reference)

	if inputSrc == image.UnknownSource {
		e := cberr.NewError(cberr.SBOMGenerationErr, "The input is not a valid scan source", err)
		logrus.Errorln(e.Error())

		return image.UnknownSource, e
	}

	return inputSrc, nil
}

// TODO: Remove with layer generation refactoring
// copyImgToLocalTarForLayers creates a local copy of the image in Docker .tarball format
// the image's metadata is converted to Docker Schema V2 in the process
func copyImgToLocalTarForLayers(src image.Source, input string, opts Option, destPathToTar string) (imagetype.Image, error) {
	var srcToUse string
	switch src {
	case image.DockerTarballSource:
		srcToUse = fmt.Sprintf("docker-archive:%s", input)
	case image.DockerDaemonSource, image.PodmanDaemonSource:
		srcToUse = fmt.Sprintf("docker-daemon:%s", input)
	case image.OciTarballSource:
		srcToUse = fmt.Sprintf("oci-archive:%s", input)
	case image.OciRegistrySource:
		srcToUse = fmt.Sprintf("docker://%s", input)
	case image.OciDirectorySource, image.UnknownSource:
		return nil, errors.New("invalid scan source")
	}

	imgSystemContext := makeDefaultSystemPullContext()

	srcRef, err := alltransports.ParseImageName(srcToUse)
	if err != nil {
		return nil, err
	}

	if opts.Credential != "" {
		username, password := opts.parseAuth()
		imgSystemContext.DockerAuthConfig = &imagetype.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	destRef, err := alltransports.ParseImageName("docker-archive:" + destPathToTar)
	if err != nil {
		return nil, err
	}

	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	policyContext, _ := signature.NewPolicyContext(policy)

	// Note: The copied manifest return here can differ slightly if the source image uses legacy formats
	// This is probably since it is converted internally by the library to a "canonical format"
	// (observed changes with docker Schema v1 images at least, where layers were changed from .tar.gzip to just .tar, which breaks the manifest digest)
	// "Sourcing" the image after copying seems to fix this and is consistent with what stereoscope reads
	_, err = copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		SourceCtx: imgSystemContext,
	})
	if err != nil {
		if errdefs.IsNotFound(err) && (src == image.DockerDaemonSource || src == image.PodmanDaemonSource) {
			// If we didn't find the image locally; retry using the registry directly with the same input
			return copyImgToLocalTarForLayers(image.OciRegistrySource, input, opts, destPathToTar)
		}
		return nil, err
	}

	// We source the image again from the tar now to avoid any problems in manifest conversion during copying
	finalTarRef, err := alltransports.ParseImageName(fmt.Sprintf("docker-archive:%s", destPathToTar))
	if err != nil {
		return nil, err
	}
	tarSrc, err := finalTarRef.NewImageSource(context.Background(), imgSystemContext)
	if err != nil {
		return nil, err
	}

	defer func(src imagetype.ImageSource) {
		_ = src.Close()
	}(tarSrc)

	img, err := containersimage.FromUnparsedImage(context.Background(), imgSystemContext, containersimage.UnparsedInstance(tarSrc, nil))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func makeDefaultSystemPullContext() *imagetype.SystemContext {
	return &imagetype.SystemContext{
		// if a multi-arch image detected, pull the linux image by default
		ArchitectureChoice:          "amd64",
		OSChoice:                    "linux",
		DockerInsecureSkipTLSVerify: imagetype.OptionalBoolTrue,
	}
}
