package scan

import (
	"crypto/sha256"
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	repo = "repo"
	tag  = "tag"

	defaultDomain    = "docker.io"
	officialRepoName = "library"
	defaultTag       = "latest"
)

var (
	digest                 = fmt.Sprintf("%x", sha256.Sum256([]byte("digest")))
	validFullTag           = fmt.Sprintf("%s/%s/%s:%s", defaultDomain, officialRepoName, repo, defaultTag)
	digestedFullTag        = fmt.Sprintf("%s/%s/%s%s%v", defaultDomain, officialRepoName, repo, digestStart, digest)
	anchoreDigestedFullTag = fmt.Sprintf("%s/%s/%s%s%v", defaultDomain, officialRepoName, repo, digestToTag, digest)

	testImageAlpineTar         = "../../test/images/alpine/alpine_3.13.5.tar"
	testImageLayersAndFilesTar = "../../test/images/files_and_layers/files_and_layers.tar"
)

func TestGenerateBomOk(t *testing.T) {
	convey.Convey("BOM generation wil valid input", t, func() {
		registryHandler := NewRegistryHandler()
		img, err := registryHandler.LoadImage(testImageAlpineTar, Option{})
		convey.So(err, convey.ShouldBeNil)

		bom, err := GenerateSBOMFromImage(img, testImageAlpineTar, "")
		convey.So(err, convey.ShouldBeNil)
		convey.So(bom, convey.ShouldNotBeNil)
	})
}

func TestBomTagFunctions(t *testing.T) {
	convey.Convey("addDefaultValuesToFullTag(tag string) (string, error)", t, func() {
		convey.Convey("assert image tag normalized", func() {
			originTag, err := addDefaultValuesToFullTag(repo)
			convey.So(err, convey.ShouldBeNil)
			wantedTag := fmt.Sprintf("%s/%s/%s:%s", defaultDomain, officialRepoName, repo, defaultTag)
			convey.So(originTag, convey.ShouldEqual, wantedTag)
		})

		convey.Convey("assert image tag normalized fail on bad tag", func() {
			originTag, err := addDefaultValuesToFullTag(fmt.Sprintf("%s:123+123", repo))
			convey.So(err, convey.ShouldBeError)
			convey.So(originTag, convey.ShouldEqual, originTag)
		})
	})

	convey.Convey("formatTag(tag string) (string, error)", t, func() {
		convey.Convey("assert digested image get a tag and normalized", func() {
			originTag, err := formatTag(repo + digestStart + digest)
			convey.So(err, convey.ShouldBeNil)
			wantedTag := fmt.Sprintf("%s/%s/%s", defaultDomain, officialRepoName, repo+digestToTag+digest)
			convey.So(originTag, convey.ShouldEqual, wantedTag)
		})

		convey.Convey("assert image tag normalized", func() {
			originTag, err := formatTag(repo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(originTag, convey.ShouldEqual, validFullTag)
		})

		convey.Convey("assert fail when on bad tag", func() {
			originTag, err := formatTag(fmt.Sprintf("%s:123+123", repo))
			convey.So(err, convey.ShouldBeError)
			convey.So(originTag, convey.ShouldEqual, originTag)
		})
	})

	convey.Convey("formatTags(tags []string, opts Option) []string", t, func() {
		convey.Convey("assert adding valid fullTag", func() {
			tags := formatTags([]string{}, validFullTag)
			convey.So(tags, convey.ShouldHaveLength, 1)
			convey.So(tags[0], convey.ShouldEqual, validFullTag)
		})

		convey.Convey("assert formatting tags (include add full tag)", func() {
			tags := formatTags([]string{digestedFullTag}, repo)
			convey.So(tags, convey.ShouldHaveLength, 2)
			convey.So(tags[0], convey.ShouldEqual, anchoreDigestedFullTag)
			// add full tag is last
			convey.So(tags[1], convey.ShouldEqual, validFullTag)
		})

		convey.Convey("assert removing only invalid tags (include add full tag)", func() {
			tags := formatTags([]string{validFullTag, "bad+sd:+"}, repo+":+12")
			convey.So(tags, convey.ShouldHaveLength, 1)
			convey.So(tags[0], convey.ShouldEqual, validFullTag)
		})
	})

	convey.Convey("revertAnchoreDigestChange(anchorNormalizedTag string) string", t, func() {
		convey.Convey("assert not affecting nun digested tags", func() {
			result := revertAnchoreDigestChange(validFullTag)
			convey.So(result, convey.ShouldEqual, validFullTag)
		})

		convey.Convey("assert reverting anchore digested tag digested tags", func() {
			result := revertAnchoreDigestChange(anchoreDigestedFullTag)
			convey.So(result, convey.ShouldEqual, digestedFullTag)
		})
	})

	convey.Convey("generateFullTagFromOriginInput(tag string, target bom.JSONImageSource) string", t, func() {
		convey.Convey("assert creation of tag from tar", func() {
			result := generateFullTagFromOriginInput(fmt.Sprintf("%s.tar", repo), digest)
			wantedTag := fmt.Sprintf("%s/%s/%s:%s", defaultDomain, officialRepoName, repo, digest)
			convey.So(result, convey.ShouldEqual, wantedTag)
		})

		convey.Convey("assert return of tag when fail adding default values", func() {
			invalidFullTag := validFullTag + "+1"
			result := generateFullTagFromOriginInput(invalidFullTag, "")
			convey.So(result, convey.ShouldEqual, invalidFullTag)
		})

		convey.Convey("assert normalized all images", func() {
			convey.Convey("digested images", func() {
				digestedImage := repo + digestStart + digest
				result := generateFullTagFromOriginInput(digestedImage, "")
				convey.So(result, convey.ShouldEqual, digestedFullTag)
			})
			convey.Convey("tagged images", func() {
				taggedImage := repo + ":" + tag
				result := generateFullTagFromOriginInput(taggedImage, "")
				convey.So(result, convey.ShouldEqual, fmt.Sprintf("%s/%s/%s", defaultDomain, officialRepoName, taggedImage))
			})
			convey.Convey("tag-less images", func() {
				taggedImage := repo
				result := generateFullTagFromOriginInput(taggedImage, "")
				convey.So(result, convey.ShouldEqual, validFullTag)
			})
		})
	})
}
