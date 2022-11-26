package eventhandler

import (
	"github.com/gookit/color"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui/component/frame"
)

// CollectLayersHandler generates a spinner while reading the layers of an image
func (h *Handler) CollectLayersHandler(line *frame.Line, value interface{}) error {
	pendingMsg := color.Bold.Sprint("Collecting layers")
	completedMsg := color.Bold.Sprint("Collected all layers")

	return h.renderStatusString(pendingMsg, completedMsg, false, true, line, value)
}
