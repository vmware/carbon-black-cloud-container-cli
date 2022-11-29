package scan

import (
	"context"
	"fmt"
	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
)

const (
	digestToTag   = ":sha256_"
	digestStart   = "@sha256:"
	tarFileEnding = ".tar"
)

// RegistryHandler coordinates with OCI registry APIs in order to retrieve
// container images as needed.
type RegistryHandler struct{}

// NewRegistryHandler constructs a new RegistryHandler instance.
func NewRegistryHandler() RegistryHandler {
	return RegistryHandler{}
}

// LoadImage parses the provided input and attempts to load an image from it
// If successful, the first returned parameter will be populated and ready to use in scanning methods
// Sharing the returned image for reading is expected, shared writing is not supported
func (h *RegistryHandler) LoadImage(input string, opts Option) (*image.Image, error) {
	logrus.WithField("input", input).Info("Loading image from source")
	src, err := parseSourceFromRawRef(input)
	if err != nil {
		logrus.WithField("input", input).WithError(err).Error("Invalid source for input, cannot load image")
		return nil, err
	}
	logrus.Debugf("Detected source is (%v)", src)

	stereoPullOptions := getImageLoadOptions()
	if opts.Credential != "" {
		logrus.Debug("Credentials passed in options, will use them to authenticate with registry")
		user, pass := opts.parseAuth()
		stereoPullOptions = append(stereoPullOptions, stereoscope.WithCredentials(image.RegistryCredentials{
			Authority: "",
			Username:  user,
			Password:  pass,
			Token:     "",
		}))
	}

	// BypassDockerDaemon only makes sense if we actually were going to pull from a daemon; ignore it for tars and similar
	if opts.BypassDockerDaemon && (src == image.DockerDaemonSource || src == image.PodmanDaemonSource) {
		logrus.WithField("input", input).Debugf("BypassDockerDaemon enabled and source is (%v), attempting to pull from registry instead", src)
		// Attempt to load directly from docker. If that fails; we fallback to loading from the daemon for backwards compatibility
		if img, err := stereoscope.GetImageFromSource(context.Background(), input, image.OciRegistrySource, stereoPullOptions...); err != nil {
			logrus.WithError(err).Warnf("Failed to pull directly from registry for input (%s); will fallback to local daemon", input)
		} else {
			logrus.WithFields(logrus.Fields{"original-input": input, "detected-source": src, "actual-source": image.OciRegistrySource}).
				Info("bypassed original source and loaded directly from registry")
			return img, nil
		}
	}

	img, err := stereoscope.GetImageFromSource(context.Background(), input, src, stereoPullOptions...)
	if err != nil {
		logrus.WithField("input", input).Error("Failed to load image from source")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"User input":         input,
		"image ID":           img.Metadata.ID,
		"Image size MB":      float64(img.Metadata.Size) / (1024 * 1024),
		"Image OS":           img.Metadata.OS,
		"Image architecture": img.Metadata.Architecture,
		"ManifestDigest":     img.Metadata.ManifestDigest,
		"RepoDigests":        img.Metadata.RepoDigests,
	}).Info("Successfully loaded image from source")
	return img, nil
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

func getImageLoadOptions() []stereoscope.Option {
	return []stereoscope.Option{
		stereoscope.WithRegistryOptions(image.RegistryOptions{
			Platform:        "linux/amd64",
			InsecureUseHTTP: true,
		}),
	}
}
