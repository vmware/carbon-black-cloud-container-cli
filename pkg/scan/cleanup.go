package scan

import (
	"github.com/anchore/stereoscope"
)

// Cleanup performs a full cleanup of temporary files related to the image pull and scan process
// It should be called after all image processing is done since instances of stereoscope/pkg/image.Image will be broken afterwards
func Cleanup() {
	// This is required mostly because of https://github.com/anchore/stereoscope/issues/132
	// It leads to hanging *.tar file every time a pull from the daemon is performed, which can fill-up /tmp over time
	// If the upstream issue is fixed, this clean up should not be needed and calling image.Cleanup() will be enough - so we can remove it

	// Note that image.Cleanup() is the better option but the two are not in conflict
	// image.Cleanup() won't panic if it runs after this code
	// stereoscope.Cleanup() will also not panic if images have been deleted successfully
	// And this call will cleanup any files left by not calling an image.Cleanup() method
	// However, all image instances will be broken after this line since their underlying files no longer exist
	stereoscope.Cleanup()
}
