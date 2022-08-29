package layers

import (
	"github.com/smartystreets/goconvey/convey"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateLayers(t *testing.T) {
	convey.Convey("assert image config", t, func() {
		convey.Convey("assert NewImageConfig", func() {
			configBytes, err := getFakeConfig()
			convey.So(err, convey.ShouldBeNil)

			imageConfig, err := NewImageConfig(configBytes)
			convey.So(err, convey.ShouldBeNil)

			convey.So(len(imageConfig.RootFs.DiffIds), convey.ShouldEqual, 1)
			convey.So(imageConfig.RootFs.DiffIds[0], convey.ShouldEqual, "sha256:b2d5eeeaba3a22b9b8aa97261957974a6bd65274ebd43e1d81d0a7b8b752b116")

			convey.So(len(imageConfig.History), convey.ShouldEqual, 2)
			convey.So(imageConfig.History[0].CreatedBy, convey.ShouldEqual, "/bin/sh -c #(nop) ADD file:8ec69d882e7f29f0652d537557160e638168550f738d0d49f90a7ef96bf31787 in / ")
			convey.So(imageConfig.History[0].EmptyLayer, convey.ShouldBeFalse)
			convey.So(imageConfig.History[0].ID, convey.ShouldEqual, "sha256:b2d5eeeaba3a22b9b8aa97261957974a6bd65274ebd43e1d81d0a7b8b752b116")

			convey.So(imageConfig.History[1].CreatedBy, convey.ShouldEqual, "/bin/sh -c #(nop)  CMD [\"/bin/sh\"]")
			convey.So(imageConfig.History[1].EmptyLayer, convey.ShouldBeTrue)
		})
	})
}

func getFakeConfig() ([]byte, error) {
	// fetch the test file position and split out the base dir
	_, testPath, _, _ := runtime.Caller(0) // nolint:dogsled
	baseDir := strings.Split(testPath, "pkg")[0]

	// get the position of the test tar file
	testFixture := filepath.Join(baseDir, "test/imageconfig/alpine_3.13.5.json")

	f, err := os.Open(testFixture)
	if err != nil {
		return nil, err
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	config, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return config, nil
}

