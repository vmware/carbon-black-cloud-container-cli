package scan

import (
	"fmt"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/source"
	"github.com/containers/image/v5/docker/reference"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
	"strings"
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

// GenerateSBOMFromImage runs the image through syft's catalogers and returns a populated SBOM of the found packages
func GenerateSBOMFromImage(img *image.Image, originalInput, forceFullTag string) (*Bom, error) {
	theSource, err := source.NewFromImage(img, originalInput)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create image source from input %v", originalInput)

		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())
		return nil, e
	}

	cft := cataloger.DefaultConfig()

	theCatalog, _, linuxDistro, err := syft.CatalogPackages(&theSource, cft)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to catalog input %v", originalInput)

		e := cberr.NewError(cberr.SBOMGenerationErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	doc, err := bom.NewJSONDocument(theCatalog, theSource.Metadata, linuxDistro, source.SquashedScope)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse the sbom for %v", originalInput)
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

// formatTag try to format the tags that Syft stores at SBOM.
// adding default tag plus repo to tags or editing digested tag.
// anchore cant handle tags with @ (they need all images to be tag and not digested).
func formatTag(tag string) (string, error) {
	if strings.Contains(tag, digestStart) {
		tag = strings.ReplaceAll(tag, digestStart, digestToTag)
	}

	return addDefaultValuesToFullTag(tag)
}

// formatTags format the attachment tags and add opts fullTag if exists.
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

func addDefaultValuesToFullTag(tag string) (string, error) {
	ref, err := reference.ParseDockerRef(tag)
	if err != nil {
		return tag, err
	}

	return ref.String(), err
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
