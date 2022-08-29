package scan

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGenerateLayers(t *testing.T) {
	convey.Convey("assert generate layers", t, func() {
		convey.Convey("assert generated layers from alpine tar", func() {
			registryHandler := RegistryHandler{registryImageCopy: mockCopyImage}
			input := "../../test/tarfile/alpine_3.13.5.tar" // TODO: improve this to be more robust

			imageLayers, err := registryHandler.GenerateLayers(input, Option{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(imageLayers, convey.ShouldNotBeNil)
			convey.So(imageLayers, convey.ShouldNotBeEmpty)

			convey.So(len(imageLayers), convey.ShouldEqual, 2)

			expectedBaseLayerDigest := "sha256:b2d5eeeaba3a22b9b8aa97261957974a6bd65274ebd43e1d81d0a7b8b752b116"
			expectedBaseLayerCMD := "#(nop) ADD file:8ec69d882e7f29f0652d537557160e638168550f738d0d49f90a7ef96bf31787 in / "

			convey.So(imageLayers[0].Digest, convey.ShouldEqual, expectedBaseLayerDigest)
			convey.So(imageLayers[0].Command, convey.ShouldEqual, expectedBaseLayerCMD)
			convey.So(imageLayers[0].Index, convey.ShouldEqual, 0)
			convey.So(imageLayers[0].Size, convey.ShouldEqual, 5608905)
			convey.So(len(imageLayers[0].Files), convey.ShouldEqual, 17)
			for _, file := range imageLayers[0].Files {
				if file.Path == "bin/busybox" {
					convey.So(file.Digest, convey.ShouldEqual, "01a52a68a05cc2bc2a115d5685be629c20cacb9f018766e365051598a0ca9bb6")
				}
				convey.So(len(file.Digest), convey.ShouldEqual, 64) // SHA256 length
				convey.So(file.Path, convey.ShouldNotEqual, "")
				convey.So(file.Size, convey.ShouldNotEqual, 0)
			}

			// Second layer is an empty layer that sets the CMD of the image -> validate that we utilize the image's manifest digest + ix
			convey.So(imageLayers[1].Digest, convey.ShouldEqual, fmt.Sprintf("%s_%d", "da67fd4f37edac865097deafd160999543243f09e055c508d369ce3ff72e6a87", 1))
			convey.So(imageLayers[1].Command, convey.ShouldEqual, "#(nop)  CMD [\"/bin/sh\"]")
			convey.So(imageLayers[1].Index, convey.ShouldEqual, 1)
			convey.So(imageLayers[1].Size, convey.ShouldEqual, 0)
			convey.So(len(imageLayers[1].Files), convey.ShouldEqual, 0)
		})
	})
}
