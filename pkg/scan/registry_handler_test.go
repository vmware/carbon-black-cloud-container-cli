package scan

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestLoadImageBadInput(t *testing.T) {
	convey.Convey("Load image with invalid input", t, func() {
		registryHandler := NewRegistryHandler()

		convey.Convey("with unknown scheme", func() {
			img, err := registryHandler.LoadImage("foo://latest", Option{})
			convey.So(img, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("with local path to a non-existent file", func() {
			img, err := registryHandler.LoadImage("/foo/bar/baz", Option{})
			convey.So(img, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("unable to expand home directory", func() {
			img, err := registryHandler.LoadImage("~foo", Option{})
			convey.So(img, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("unknown source", func() {
			img, err := registryHandler.LoadImage("foo:bar:baz", Option{})
			convey.So(img, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
