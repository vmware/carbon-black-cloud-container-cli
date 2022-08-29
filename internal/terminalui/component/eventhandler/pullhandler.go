package eventhandler

import (
	"fmt"
	"strings"
	"time"

	"github.com/anchore/stereoscope/pkg/image/docker"
	humanize "github.com/dustin/go-humanize"
	"github.com/gookit/color"
	progress "github.com/wagoodman/go-progress"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/frame"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui/component/spinner"
)

var (
	dockerPullCompletedColor = color.HEX("#fcba03")
	dockerPullDownloadColor  = color.HEX("#777777")
	dockerPullExtractColor   = color.White
	dockerPullStageChars     = strings.Split("▁▃▄▅▆▇█", "")
)

// PullDockerImageHandler periodically writes a formatted line widget representing a docker image pull event.
// This function is modified based on 'github.com/anchore/syft/ui/event_handlers.PullDockerImageHandler'.
func (h *Handler) PullDockerImageHandler(line *frame.Line, value interface{}) error {
	pullStatus := value.(*docker.PullStatus)

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()

		s := spinner.NewSpinnerWithCharset(spinner.DefaultDotSet)

	loop:
		for {
			select {
			case <-h.ctx.Done():
				break loop
			case <-time.After(interval):
				formatDockerImagePullStatus(pullStatus, s, line)
				if pullStatus.Complete() {
					break loop
				}
			}
		}

		if pullStatus.Complete() {
			title := color.Bold.Sprint("Pulled image")
			_ = line.Render(fmt.Sprintf(statusTitleTemplate, completedStatus, title))
		}
	}()

	return nil
}

// formatDockerImagePullStatus writes the docker image pull status summarized into a single line for the given state.
func formatDockerImagePullStatus(pullStatus *docker.PullStatus, s *spinner.Spinner, line *frame.Line) {
	var size, current uint64

	title := color.Bold.Sprint("Pulling image")

	layers := pullStatus.Layers()
	status := make(map[docker.LayerID]docker.LayerState)
	completed := make([]string, len(layers))

	// fetch the current state
	for idx, layer := range layers {
		completed[idx] = " "
		status[layer] = pullStatus.Current(layer)
	}

	numCompleted := 0

	for idx, layer := range layers {
		prog := status[layer].PhaseProgress
		progCurrent := prog.Current()
		progSize := prog.Size()

		if progress.IsCompleted(prog) {
			input := dockerPullStageChars[len(dockerPullStageChars)-1]
			completed[idx] = formatDockerPullPhase(status[layer].Phase, input)
		} else if progCurrent != 0 {
			var ratio float64
			switch {
			case progCurrent == 0 || progSize < 0:
				ratio = 0
			case progCurrent >= progSize:
				ratio = 1
			default:
				ratio = float64(progCurrent) / float64(progSize)
			}

			i := int(ratio * float64(len(dockerPullStageChars)-1))
			input := dockerPullStageChars[i]
			completed[idx] = formatDockerPullPhase(status[layer].Phase, input)
		}

		if progress.IsErrCompleted(status[layer].DownloadProgress.Error()) {
			numCompleted++
		}
	}

	for _, layer := range layers {
		prog := status[layer].DownloadProgress
		size += uint64(prog.Size())
		current += uint64(prog.Current())
	}

	var progStr, auxInfo string

	if len(layers) > 0 {
		render := strings.Join(completed, "")
		prefix := dockerPullCompletedColor.Sprintf("%d layers", len(layers))
		auxInfo = auxInfoFormat.Sprintf("[%s / %s]", humanize.Bytes(current), humanize.Bytes(size))

		if len(layers) == numCompleted {
			auxInfo = auxInfoFormat.Sprintf("[%s] [extracting]", humanize.Bytes(size))
		}

		progStr = fmt.Sprintf("%s▕%s▏", prefix, render)
	}

	spin := color.Magenta.Sprint(s.Next())
	_ = line.Render(fmt.Sprintf(statusTitleTemplate+"%s%s", spin, title, progStr, auxInfo))
}

// formatDockerPullPhase returns a single character that represents the status of a layer pull.
func formatDockerPullPhase(phase docker.PullPhase, inputStr string) string {
	switch phase {
	case docker.WaitingPhase:
		// ignore any progress related to waiting
		return " "
	case docker.PullingFsPhase, docker.DownloadingPhase:
		return dockerPullDownloadColor.Sprint(inputStr)
	case docker.DownloadCompletePhase:
		return dockerPullDownloadColor.Sprint(dockerPullStageChars[len(dockerPullStageChars)-1])
	case docker.ExtractingPhase:
		return dockerPullExtractColor.Sprint(inputStr)
	case docker.VerifyingChecksumPhase, docker.PullCompletePhase:
		return dockerPullCompletedColor.Sprint(inputStr)
	case docker.AlreadyExistsPhase:
		return dockerPullCompletedColor.Sprint(dockerPullStageChars[len(dockerPullStageChars)-1])
	case docker.UnknownPhase:
		fallthrough
	default:
		return inputStr
	}
}
