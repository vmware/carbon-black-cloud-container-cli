// Package dynamicui provides display handler for dynamic progress bar in the terminal
package dynamicui

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/eventhandler"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/frame"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
	"gitlab.bit9.local/octarine/cbctl/pkg/presenter"
)

// Display will help us handle all the incoming events and show them on the terminal.
type Display struct{}

// NewDisplay will initialize a display instance.
func NewDisplay() *Display {
	return &Display{}
}

// displayResults will display result on terminal.
func displayResults(errorMsg string, fr *frame.Frame, wg *sync.WaitGroup, e bus.Event) error {
	var displayErrLocal error

	wg.Wait()

	var pres presenter.Presenter

	pres, ok := e.Value().(presenter.Presenter)
	if !ok {
		return fmt.Errorf("%v internal error", errorMsg)
	}

	fr.Append()
	displayErrLocal = fr.Append().Render(color.Bold.Sprint(pres.Title()))
	fr.Append()

	if err := pres.Present(os.Stdout); err != nil {
		displayErrLocal = fmt.Errorf("%v %v", errorMsg, err)
	}

	if pres.Footer() != "" {
		displayErrLocal = fr.Append().Render(color.Bold.Sprint(pres.Footer()))
	}

	return displayErrLocal
}

// DisplayEvents will read events from channel, and show them on terminal.
func (d Display) DisplayEvents() {
	var (
		displayErr error
		exitCode   = 0
	)

	fr := frame.NewFrame(os.Stderr)
	_ = fr.HideCursor()

	defer func() {
		fr.Append()
		_ = fr.ShowCursor()

		if displayErr != nil {
			msg := "Failed to show the ui during the whole process"
			e := cberr.NewError(cberr.DisplayErr, msg, displayErr)
			_, _ = fmt.Fprintln(os.Stderr, msg)
			exitCode = e.ExitCode()

			logrus.Errorln(e)
		}

		if exitCode > 0 {
			os.Exit(exitCode)
		}
	}()

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	handler := eventhandler.NewHandler(ctx, wg)

eventLoop:
	for e := range bus.EventChan() {
		switch e.Type() {
		case bus.NewVersionAvailable:
			msg := color.Magenta.Sprint(e.Value())
			displayErr = fr.Append().Render(msg)
		case bus.NewMessageDetected, bus.ValidateFinishedSuccessfully:
			wg.Wait()
			msg := color.Bold.Sprint(e.Value())
			displayErr = fr.Append().Render(msg)
		case bus.NewErrorDetected:
			msg := fmt.Sprintf("%s %v", color.Red.Sprint("[Error]"), e.Value())
			displayErr = fr.Append().Render(msg)
			exitCode = e.(*bus.ErrorEvent).ExitCode()
		case bus.PullDockerImage:
			displayErr = handler.PullDockerImageHandler(fr.Append(), e.Value())
		case bus.CopyImage:
			displayErr = handler.CopyImageHandler(fr.Append(), e.Value())
		case bus.ReadImage:
			displayErr = handler.ReadImageHandler(fr.Append(), e.Value())
		case bus.FetchImage:
			displayErr = handler.FetchImageHandler(fr.Append(), e.Value())
		case bus.CatalogerStarted:
			displayErr = handler.CatalogerStartedHandler(fr.Append(), e.Value())
		case bus.ScanStarted:
			displayErr = handler.AnalyzeStartedHandler(fr.Append(), e.Value())
		case bus.ScanFinished, bus.ValidateFinishedWithViolations:
			errorMsg := "failed to show vulnerability results:"
			displayErr = displayResults(errorMsg, fr, wg, e)
		case bus.PrintSBOM:
			errorMsg := "failed to show packages:"
			displayErr = displayResults(errorMsg, fr, wg, e)
		case bus.CatalogerFinished, bus.ReadLayer:
			fallthrough
		default:
			continue
		}

		if e.IsEnd() || displayErr != nil {
			break eventLoop
		}
	}
}
