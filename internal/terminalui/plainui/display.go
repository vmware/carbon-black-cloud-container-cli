// Package plainui provides display handler for generic terminal
package plainui

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
	"gitlab.bit9.local/octarine/cbctl/pkg/presenter"
)

// Display will help us handle all the incoming events and show them on the terminal.
type Display struct{}

// NewDisplay will initialize a display instance.
func NewDisplay() *Display {
	return &Display{}
}

// displayResults will display result.
func displayResults(e bus.Event) error {
	var displayErrLocal error

	var pres presenter.Presenter

	pres, ok := e.Value().(presenter.Presenter)
	if !ok {
		return fmt.Errorf("internal error in display results")
	}

	_ = printMessageOnStderr("")
	displayErrLocal = printMessageOnStderr(pres.Title())

	if err := pres.Present(os.Stdout); err != nil {
		displayErrLocal = fmt.Errorf("failed to show results: %v", err)
	}

	if pres.Footer() != "" {
		displayErrLocal = printMessageOnStderr(pres.Footer())
	}

	return displayErrLocal
}

// DisplayEvents will read events from channel, and show them on terminal.
func (d Display) DisplayEvents() {
	var (
		displayErr error
		exitCode   = 0
	)

	defer func() {
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

eventLoop:
	for e := range bus.EventChan() {
		switch e.Type() {
		case bus.NewVersionAvailable:
			displayErr = printMessageOnStderr(e.Value())
		case bus.NewMessageDetected, bus.ValidateFinishedSuccessfully:
			displayErr = printMessageOnStderr(e.Value())
		case bus.NewErrorDetected:
			msg := fmt.Sprintf("[Error] %v", e.Value())
			displayErr = printMessageOnStderr(msg)
			exitCode = e.(*bus.ErrorEvent).ExitCode()
		case bus.PullDockerImage:
			msg := "Pulling Docker image..."
			displayErr = printMessageOnStderr(msg)
		case bus.CopyImage:
			msg := "Copying image..."
			displayErr = printMessageOnStderr(msg)
		case bus.ReadImage:
			msg := "Reading image..."
			displayErr = printMessageOnStderr(msg)
		case bus.FetchImage:
			msg := "Fetching image..."
			displayErr = printMessageOnStderr(msg)
		case bus.CatalogerStarted:
			msg := "Starting cataloger..."
			displayErr = printMessageOnStderr(msg)
		case bus.ScanStarted:
			msg := "Analyzing image..."
			displayErr = printMessageOnStderr(msg)
		case bus.ScanFinished, bus.ValidateFinishedWithViolations:
			displayErr = displayResults(e)
		case bus.PrintSBOM:
			displayErr = displayResults(e)
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

// printMessageOnStderr will print the message on stderr.
func printMessageOnStderr(msg interface{}) error {
	if _, err := fmt.Fprintln(os.Stderr, msg); err != nil {
		return err
	}

	return nil
}
