/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package eventhandler

import (
	"fmt"

	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/gookit/color"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui/component/frame"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui/component/spinner"
	"github.com/wagoodman/go-progress"
)

// CatalogerStartedHandler periodically writes catalog statistics to a single line.
func (h *Handler) CatalogerStartedHandler(line *frame.Line, m interface{}) error {
	monitor := m.(cataloger.Monitor)

	h.wg.Add(1)

	stream := progress.StreamMonitors(
		h.ctx, []progress.Monitorable{monitor.FilesProcessed, monitor.PackagesDiscovered}, interval)
	title := color.Bold.Sprint("Cataloging image")

	go func() {
		defer h.wg.Done()

		s := spinner.NewSpinnerWithCharset(spinner.DefaultDotSet)

		for p := range stream {
			nextSpinner := color.Magenta.Sprint(s.Next())
			auxInfo := auxInfoFormat.Sprintf("[packages %d]", p[1])
			_ = line.Render(fmt.Sprintf(statusTitleTemplate+"%s", nextSpinner, title, auxInfo))
		}

		title = color.Bold.Sprint("Cataloged image")
		auxInfo := auxInfoFormat.Sprintf("[%d packages]", monitor.PackagesDiscovered.Current())
		_ = line.Render(fmt.Sprintf(statusTitleTemplate+"%s", completedStatus, title, auxInfo))
	}()

	return nil
}
