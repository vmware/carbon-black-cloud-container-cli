package eventhandler

import (
	"github.com/gookit/color"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui/component/frame"
)

// StartScanHandler generates a spinner while waiting for image id
func (h *Handler) StartScanHandler(line *frame.Line, value interface{}) error {
	pendingMsg := color.Bold.Sprint("Scan in progress")
	completedMsg := color.Bold.Sprint("Scan in progress")

	return h.renderStatusString(pendingMsg, completedMsg, false, true, line, value)
}
