package scan

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
	"testing"
)

func TestGenerateLayers(t *testing.T) {
	convey.Convey("assert generate layers", t, func() {
		convey.Convey("from alpine tar", func() {
			registryHandler := NewRegistryHandler()
			input := testImageAlpineTar

			img, err := registryHandler.LoadImage(input, Option{})
			convey.So(err, convey.ShouldBeNil)
			imageLayers, err := GenerateLayersAndFileData(img)
			convey.So(err, convey.ShouldBeNil)
			convey.So(imageLayers, convey.ShouldNotBeNil)
			convey.So(imageLayers, convey.ShouldNotBeEmpty)

			convey.So(len(imageLayers), convey.ShouldEqual, 2)

			expectedBaseLayerDigest := "sha256:b2d5eeeaba3a22b9b8aa97261957974a6bd65274ebd43e1d81d0a7b8b752b116"
			expectedBaseLayerCMD := "/bin/sh -c #(nop) ADD file:8ec69d882e7f29f0652d537557160e638168550f738d0d49f90a7ef96bf31787 in / "

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
			convey.So(imageLayers[1].Digest, convey.ShouldEqual, fmt.Sprintf("%s_%d", "sha256:da67fd4f37edac865097deafd160999543243f09e055c508d369ce3ff72e6a87", 1))
			convey.So(imageLayers[1].Command, convey.ShouldEqual, "/bin/sh -c #(nop)  CMD [\"/bin/sh\"]")
			convey.So(imageLayers[1].Index, convey.ShouldEqual, 1)
			convey.So(imageLayers[1].Size, convey.ShouldEqual, 0)
			convey.So(len(imageLayers[1].Files), convey.ShouldEqual, 0)
		})

		convey.Convey("from our own dummy image", func() {
			registryHandler := NewRegistryHandler()
			input := testImageLayersAndFilesTar

			img, err := registryHandler.LoadImage(input, Option{})
			convey.So(err, convey.ShouldBeNil)

			imageLayers, err := GenerateLayersAndFileData(img)
			convey.So(err, convey.ShouldBeNil)
			convey.So(imageLayers, convey.ShouldNotBeEmpty)

			// See the dockerfile for this image under tests/images/.. if the below doesn't make sense
			// Note that changing the binary and rebuilding the whole image can change the digests below
			binaryDigest := "20e96e1ed34f1b144ee3514421b8faf4e421ec3b36661c9dd35186619a44cb6c" // Change this if the binary under _app is changed in any way

			// First layer should have the binary
			l1 := imageLayers[0]
			convey.So(l1.Index, convey.ShouldEqual, 0)
			convey.So(l1.Digest, convey.ShouldEqual, "sha256:e11ffdd32b294c134aa6e6f13a791120e6b8e694cd23797522b5ae6bb144bebd")
			convey.So(l1.Command, convey.ShouldStartWith, "ADD app /executable")
			convey.So(l1.IsEmpty, convey.ShouldBeFalse)
			convey.So(l1.Files, convey.ShouldHaveLength, 1)
			convey.So(l1.Files[0].InSquashedImage, convey.ShouldBeTrue)
			convey.So(l1.Files[0].Digest, convey.ShouldEqual, binaryDigest)
			convey.So(l1.Files[0].Path, convey.ShouldEqual, "/executable")
			convey.So(l1.Files[0].Category, convey.ShouldEqual, layers.CategoryElf)

			// Second layer should have another instance of the binary that is NOT in the final image (as it's deleted)
			l2 := imageLayers[1]
			convey.So(l2.Index, convey.ShouldEqual, 1)
			convey.So(l2.Command, convey.ShouldStartWith, "ADD app /executable_2")
			convey.So(l2.IsEmpty, convey.ShouldBeFalse)
			convey.So(l2.Files, convey.ShouldHaveLength, 1)
			convey.So(l2.Files[0].Path, convey.ShouldEqual, "/executable_2")
			convey.So(l2.Files[0].Digest, convey.ShouldEqual, binaryDigest)
			convey.So(l2.Files[0].InSquashedImage, convey.ShouldBeFalse)

			// Third layer should be non-empty but not have any files - as it deletes the binary in L2 and nothing else
			l3 := imageLayers[2]
			convey.So(l3.Index, convey.ShouldEqual, 2)
			convey.So(l3.Command, convey.ShouldStartWith, "RUN /executable delete /executable_2")
			convey.So(l3.IsEmpty, convey.ShouldBeFalse)
			convey.So(l3.Files, convey.ShouldBeEmpty)

			// Fourth layer should have another instance of the binary that is NOT in the final image (as it's modified)
			l4 := imageLayers[3]
			convey.So(l4.Index, convey.ShouldEqual, 3)
			convey.So(l4.Command, convey.ShouldStartWith, "ADD app /executable_3")
			convey.So(l4.IsEmpty, convey.ShouldBeFalse)
			convey.So(l4.Files, convey.ShouldHaveLength, 1)
			convey.So(l4.Files[0].Path, convey.ShouldEqual, "/executable_3")
			convey.So(l4.Files[0].Digest, convey.ShouldEqual, binaryDigest)
			convey.So(l4.Files[0].InSquashedImage, convey.ShouldBeFalse)

			// Fifth layer should have the instances of the binary from L4 but this time it IS in the final image
			l5 := imageLayers[4]
			convey.So(l5.Index, convey.ShouldEqual, 4)
			convey.So(l5.Command, convey.ShouldStartWith, "RUN /executable change /executable_3")
			convey.So(l5.IsEmpty, convey.ShouldBeFalse)
			convey.So(l5.Files, convey.ShouldHaveLength, 1)
			convey.So(l5.Files[0].Path, convey.ShouldEqual, "/executable_3")
			convey.So(l5.Files[0].InSquashedImage, convey.ShouldBeTrue)

			// Sixth layer should be empty as it modifies env variables
			l6 := imageLayers[5]
			convey.So(l6.Index, convey.ShouldEqual, 5)
			convey.So(l6.Command, convey.ShouldStartWith, "ENV test=/test")
			convey.So(l6.IsEmpty, convey.ShouldBeTrue)
			convey.So(l6.Files, convey.ShouldBeEmpty)
		})
	})
}
