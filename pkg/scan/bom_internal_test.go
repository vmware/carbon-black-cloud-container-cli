package scan

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/smartystreets/goconvey/convey"
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
)

func TestGenerateBomOk(t *testing.T) {
	registryHandler := RegistryHandler{registryImageCopy: mockCopyImage}

	bom, err := registryHandler.Generate("foo:latest", Option{BypassDockerDaemon: true})
	if err != nil {
		t.Errorf("failed to generate BOM: %v", err)
	}

	if bom == nil {
		t.Error("BOM was unexpectedly nil")
	}
}

func TestGenerateBomBadImage(t *testing.T) {
	registryHandler := RegistryHandler{registryImageCopy: mockCopyImage}

	// Unknown scheme
	bom, err := registryHandler.Generate("foo://latest", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}

	// No such file
	bom, err = registryHandler.Generate("/foo/bar/baz", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}

	// Unable to expand home directory
	bom, err = registryHandler.Generate("~foo", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}

	// Unknown source
	bom, err = registryHandler.Generate("foo:bar:baz", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}
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

func mockCopyImage(
	ctx context.Context,
	policyContext *signature.PolicyContext,
	destRef,
	srcRef types.ImageReference,
	options *copy.Options,
) (copiedManifest []byte, retErr error) {
	// fetch the test file position and split out the base dir
	_, testPath, _, _ := runtime.Caller(0) // nolint:dogsled
	baseDir := strings.Split(testPath, "pkg")[0]

	// get the position of the test tar file
	testFixture := filepath.Join(baseDir, "test/tarfile/alpine_3.13.5.tar")

	src, err := alltransports.ParseImageName(fmt.Sprintf("docker-archive:%s", testFixture))
	if err != nil {
		return nil, err
	}

	return copy.Image(ctx, policyContext, destRef, src, options)
}
