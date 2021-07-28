/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package scan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/httptool"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/version"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/image"
	"github.com/wagoodman/go-progress"
)

const (
	service = "management/images"

	healthCheckTemplate = "%s/" + service + "/health"
	putSBOMTemplate     = "%s/" + service + "/%s"
	getStatusTemplate   = "%s/" + service + "/%s/status"
	getVulnTemplate     = "%s/" + service + "/%s/vulnerabilities"
)

// Status is the status for the scanning.
type Status string

// Detailed statuses of the scanning result.
const (
	UploadedStatus Status = "UPLOADED"
	FinishedStatus Status = "FINISHED"
	QueuedStatus   Status = "QUEUED"
	FailedStatus   Status = "FAILED"
)

// Handler has all the fields for sending request to scanning service.
type Handler struct {
	session      *httptool.RequestSession
	basePath     string
	buildStep    string
	namespace    string
	bom          *Bom
	pollPause    time.Duration
	pollInterval time.Duration
	pollDuration time.Duration
}

// analysisPayload is the payload used for uploading sbom to image scanning service.
type analysisPayload struct {
	SBOM      *bom.JSONDocument `json:"sbom"`
	BuildStep string            `json:"build_step"`
	Namespace string            `json:"namespace"`
	ForceScan bool              `json:"force_scan"`
	Meta      struct {
		SyftVersion string `json:"syft_version"`
		CliVersion  string `json:"cli_version"`
	} `json:"metadata"`
}

// NewScanHandler will create a handler for scan cmd.
func NewScanHandler(saasTmpl, orgKey, apiID, apiKey string, bom *Bom) *Handler {
	saasTmpl = strings.Trim(saasTmpl, "/")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/orgs")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1beta")

	// for image scanning service, we need to add a beta in the version for now
	// should remove when image-scanning-service not use beta version
	basePath := fmt.Sprintf("%s/v1beta/orgs/%s", saasTmpl, orgKey)
	session := httptool.NewRequestSession(apiID, apiKey)

	return newScanHandler(session, basePath, bom)
}

func newScanHandler(session *httptool.RequestSession, basePath string, bom *Bom) *Handler {
	return &Handler{
		session:      session,
		basePath:     basePath,
		bom:          bom,
		pollPause:    10 * time.Second,
		pollDuration: 600 * time.Second,
		pollInterval: 5 * time.Second,
	}
}

// AttachSBOMBuildStepAndNamespace will attach sbom & policy to the handler.
func (h *Handler) AttachSBOMBuildStepAndNamespace(bom *Bom, buildStep, namespace string) {
	h.buildStep = buildStep
	h.namespace = namespace
	h.bom = bom
}

// HealthCheck will check the health of the service backend.
func (h Handler) HealthCheck() error {
	healthCheckPath := fmt.Sprintf(healthCheckTemplate, h.basePath)
	_, _, err := h.session.RequestData(http.MethodGet, healthCheckPath, nil)

	return err
}

// Scan will send payload to image scanning service and fetch the result back.
func (h *Handler) Scan(opts Option) (*image.ScannedImage, error) {
	// update scan duration from the options
	if opts.Timeout > 0 {
		h.pollDuration = time.Duration(opts.Timeout) * time.Second
	}

	stage := &progress.Stage{}
	prog := &progress.Manual{Total: 1}
	value := progress.StagedProgressable(&struct {
		progress.Stager
		progress.Progressable
	}{
		Stager:       stage,
		Progressable: prog,
	})
	bus.Publish(bus.NewEvent(bus.ScanStarted, value, false))

	defer prog.SetCompleted()

	currentStage := "initializing scan"
	logrus.Infof("Scanning image, current stage: %v", currentStage)
	stage.Current = currentStage

	signalChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	scannedImageChan := make(chan *image.ScannedImage)

	// catch incoming signal, if it is a ^C, send a cancel notification to backend
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		currentStage = "uploading software bills of material"
		logrus.Infof("Scanning image, current stage: %v", currentStage)
		stage.Current = currentStage

		if status, err := h.PutBomToAnalysisAPI(opts); err != nil {
			errChan <- err
			return
		} else if status == FinishedStatus {
			// the result should be fetched directly from backend
			scannedImage, e := h.GetImageVulnerability(h.bom.ManifestDigest)
			if e == nil && scannedImage != nil && scannedImage.ScanStatus == "SCANNED" {
				currentStage = "fetching result"
				logrus.Infof("Scanning image, current stage: %v", currentStage)
				stage.Current = currentStage
				scannedImageChan <- scannedImage

				return
			}
		}

		currentStage = "analyzing vulnerabilities from sbom"
		logrus.Infof("Scanning image, current stage: %v", currentStage)
		stage.Current = currentStage

		// sleep for a while, since we need some time to analyze
		time.Sleep(h.pollPause)

		scannedImage, err := h.GetResponseFromScanAPI(h.bom.ManifestDigest)
		if err != nil {
			errChan <- err
			return
		}

		currentStage = "fetching result"
		logrus.Infof("Scanning image, current stage: %v", currentStage)
		stage.Current = currentStage

		scannedImageChan <- scannedImage
	}()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case <-signalChan:
			if err := h.SendCancelSignal(h.bom.ManifestDigest); err != nil {
				logrus.Errorln(err)
			}

			err := fmt.Errorf("detect ^C signal interruption")
			logrus.Errorln(err)
			errChan <- err
		case scannedImage := <-scannedImageChan:
			if scannedImage != nil {
				return scannedImage, nil
			}
		}
	}
}

// PutBomToAnalysisAPI will call the PUT API and upload sbom to image scanning service.
func (h Handler) PutBomToAnalysisAPI(opts Option) (Status, error) {
	versionInfo := version.GetCurrentVersion()
	payload := &analysisPayload{
		SBOM:      &h.bom.Packages,
		BuildStep: h.buildStep,
		Namespace: h.namespace,
		ForceScan: opts.ForceScan,
		Meta: struct {
			SyftVersion string `json:"syft_version"`
			CliVersion  string `json:"cli_version"`
		}{
			SyftVersion: versionInfo.SyftVersion,
			CliVersion:  versionInfo.Version,
		},
	}

	analysisPath := fmt.Sprintf(putSBOMTemplate, h.basePath, h.bom.ManifestDigest)
	if h.bom != nil && h.bom.FullTag != "" {
		statusURLWithQueries, _ := url.Parse(analysisPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		statusURLWithQueries.RawQuery = params.Encode()
		analysisPath = statusURLWithQueries.String()
	}

	_, resp, err := h.session.RequestData(http.MethodPut, analysisPath, payload)
	if err != nil && cberr.ErrorCode(err) == cberr.HTTPUnsuccessfulResponseErr {
		errMsg := "Failed to put sbom to the backend"
		e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
		logrus.Error(e.Error())

		return "", e
	}

	return Status(resp), err
}

// GetResponseFromScanAPI will call the status API from image scanning service periodically,
// once the status is "FINISHED", it will fetch the real result from vuln API.
func (h Handler) GetResponseFromScanAPI(digest string) (*image.ScannedImage, error) {
	ticker := time.NewTicker(h.pollInterval)
	defer ticker.Stop()

	timeout := make(chan bool)
	statusResult := make(chan Status)

	go func() {
		time.Sleep(h.pollDuration)
		timeout <- true
	}()

	for {
		select {
		case <-timeout:
			if err := h.SendCancelSignal(digest); err != nil {
				logrus.Errorln(err)
			}

			return nil, cberr.NewError(cberr.TimeoutErr, "Time out during status polling", nil)
		case <-ticker.C:
			go func() {
				if status, err := h.GetImageAnalysisStatus(digest); err == nil {
					statusResult <- status
				} else {
					statusResult <- FailedStatus
				}
			}()
		case result := <-statusResult:
			switch result {
			case FinishedStatus:
				return h.GetImageVulnerability(digest)
			case FailedStatus:
				errMsg := fmt.Sprintf("Failed to scan image [%s]", digest)
				return nil, cberr.NewError(cberr.TimeoutErr, errMsg, nil)
			case QueuedStatus, UploadedStatus:
				fallthrough
			default:
				continue
			}
		}
	}
}

// GetImageAnalysisStatus will fetch the current analysis result of an image.
func (h Handler) GetImageAnalysisStatus(digest string) (Status, error) {
	statusPath := fmt.Sprintf(getStatusTemplate, h.basePath, digest)
	if h.bom != nil && h.bom.FullTag != "" {
		statusURLWithQueries, _ := url.Parse(statusPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		statusURLWithQueries.RawQuery = params.Encode()
		statusPath = statusURLWithQueries.String()
	}

	_, resp, err := h.session.RequestData(http.MethodGet, statusPath, nil)
	if err != nil {
		// the scanning message might still be in the queue, return QUEUED here
		if cberr.ErrorCode(err) == cberr.HTTPNotFoundErr {
			return QueuedStatus, nil
		}

		if cberr.ErrorCode(err) == cberr.HTTPUnsuccessfulResponseErr {
			errMsg := fmt.Sprintf("Failed to fetch status for image (%v)", digest)
			e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
			logrus.Errorln(e.Error())

			return "", e
		}

		return "", err
	}

	return Status(resp), nil
}

// GetImageVulnerability will fetch the vulnerability result via image digest.
func (h Handler) GetImageVulnerability(digest string) (*image.ScannedImage, error) {
	vulnPath := fmt.Sprintf(getVulnTemplate, h.basePath, digest)
	if h.bom != nil && h.bom.FullTag != "" {
		vulnURLWithQueries, _ := url.Parse(vulnPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		vulnURLWithQueries.RawQuery = params.Encode()
		vulnPath = vulnURLWithQueries.String()
	}

	_, resp, err := h.session.RequestData(http.MethodGet, vulnPath, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to fetch scan report for image (%v)", digest)
		e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	var scannedImage image.ScannedImage

	if err := json.Unmarshal(resp, &scannedImage); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal scan report for image (%v)", digest)
		e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	return &scannedImage, nil
}

// SendCancelSignal will send a cancel signal to backend, will be called when timeout or manual interruption.
func (h Handler) SendCancelSignal(digest string) error {
	logrus.Info("Sending cancel signal to backend")

	statusPath := fmt.Sprintf(getStatusTemplate, h.basePath, digest)
	if h.bom != nil && h.bom.FullTag != "" {
		statusURLWithQueries, _ := url.Parse(statusPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		statusURLWithQueries.RawQuery = params.Encode()
		statusPath = statusURLWithQueries.String()
	}

	_, _, err := h.session.RequestData(http.MethodPut, statusPath, struct {
		ScanStatus string `json:"scan_status"`
	}{
		ScanStatus: "NOT_SCANNED",
	})

	return err
}
