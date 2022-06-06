package eventhandler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gookit/color"
	"github.com/wagoodman/go-progress"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/frame"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/progressformatter"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/spinner"
)

const (
	// fmt padding will also take ascii code into account, so we need to set the padding a little longer.
	statusTitleTemplate = " %s %-25s "
	barWidth            = 40
	interval            = 150 * time.Millisecond
)

var (
	completedStatus = color.Green.Sprint("⠿")
	auxInfoFormat   = color.HEX("#777777")
)

// Handler is the handler for all events.
type Handler struct {
	ctx context.Context
	wg  *sync.WaitGroup
}

// NewHandler will initialize a handler.
func NewHandler(ctx context.Context, wg *sync.WaitGroup) *Handler {
	return &Handler{
		ctx: ctx,
		wg:  wg,
	}
}

// CopyImageHandler will render a spinner and record the process for copying image.
func (h *Handler) CopyImageHandler(line *frame.Line, value interface{}) error {
	pendingMsg := color.Bold.Sprint("Copying image")
	completedMsg := color.Bold.Sprint("Copied image")

	return h.renderStatusString(pendingMsg, completedMsg, false, true, line, value)
}

// ReadImageHandler periodically writes a the image read/parse/build-tree status in the form of a progress bar.
func (h *Handler) ReadImageHandler(line *frame.Line, value interface{}) error {
	pendingMsg := color.Bold.Sprint("Parsing image")
	completedMsg := color.Bold.Sprint("Parsed image")

	return h.renderStatusString(pendingMsg, completedMsg, true, false, line, value)
}

// FetchImageHandler periodically writes a the image save and write-to-disk process in the form of a progress bar.
func (h *Handler) FetchImageHandler(line *frame.Line, value interface{}) error {
	pendingMsg := color.Bold.Sprint("Loading image")
	completedMsg := color.Bold.Sprint("Loaded image")

	return h.renderStatusString(pendingMsg, completedMsg, true, true, line, value)
}

// AnalyzeStartedHandler will generate a spinner during analyzing images.
func (h *Handler) AnalyzeStartedHandler(line *frame.Line, value interface{}) error {
	pendingMsg := color.Bold.Sprint("Scanning image")
	completedMsg := color.Bold.Sprint("Scanned image")

	return h.renderStatusString(pendingMsg, completedMsg, false, true, line, value)
}

func (h *Handler) renderStatusString(
	pendingMsg, completedMsg string, barEnabled, auxEnabled bool, line *frame.Line, value interface{},
) error {
	h.wg.Add(1)

	stream := progress.Stream(h.ctx, value.(progress.Progressable), interval)

	go func() {
		defer h.wg.Done()

		s := spinner.NewSpinnerWithCharset(spinner.DefaultDotSet)
		f := progressformatter.NewFormatterWithTheme(barWidth, progressformatter.DefaultTheme)

		nextSpinner := color.Magenta.Sprint(s.Next())
		_ = line.Render(fmt.Sprintf(statusTitleTemplate, nextSpinner, pendingMsg))

		auxInfo := ""
		progStr := ""

		for p := range stream {
			nextSpinner = color.Magenta.Sprint(s.Next())

			if auxEnabled {
				auxInfo = auxInfoFormat.Sprintf("[%s]", value.(progress.StagedProgressable).Stage())
			}

			if barEnabled {
				progStr = f.Format(p) + " "
			}

			lineMsg := fmt.Sprintf(statusTitleTemplate+"%s%s", nextSpinner, pendingMsg, progStr, auxInfo)
			_ = line.Render(lineMsg)
		}

		_ = line.Render(fmt.Sprintf(statusTitleTemplate+"%s", completedStatus, completedMsg, auxInfo))
	}()

	return nil
}
