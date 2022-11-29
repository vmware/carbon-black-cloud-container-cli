package scan

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
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
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/image"
	progress "github.com/wagoodman/go-progress"
)

const (
	service             = "%s/analyzer"
	healthCheckTemplate = service + "/health"

	imagesRoute     = service + "/images/%s"
	getVulnTemplate = imagesRoute + "/vulnerabilities"

	operationRoute    = imagesRoute + "/operations/%s"
	putSBOMTemplate   = operationRoute
	getStatusTemplate = operationRoute + "/status"

	imageIsRoute        = service + "/image_id/%s"
	getVulnByIdTemplate = imageIsRoute + "/vulnerabilities"
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

type StatusResponse struct {
	OperationStatus Status `json:"operation_status"`
}

// Handler has all the fields for sending request to scanning service.
type Handler struct {
	session      *httptool.RequestSession
	basePath     string
	buildStep    string
	namespace    string
	imageID      string
	bom          *Bom
	layers       []layers.Layer
	pollPause    time.Duration
	pollInterval time.Duration
	pollDuration time.Duration
}

// NewScanHandler will create a handler for scan cmd.
func NewScanHandler(saasTmpl, orgKey, apiID, apiKey string, bom *Bom, layers []layers.Layer) *Handler {
	saasTmpl = strings.Trim(saasTmpl, "/")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/orgs")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1beta")

	// for image scanning service, we need to add a beta in the version for now
	// should remove when image-scanning-service not use beta version
	basePath := fmt.Sprintf("%s/v1beta/orgs/%s", saasTmpl, orgKey)
	session := httptool.NewRequestSession(apiID, apiKey)

	return newScanHandler(session, basePath, bom, layers)
}

func newScanHandler(session *httptool.RequestSession, basePath string, bom *Bom, layers []layers.Layer) *Handler {
	return &Handler{
		session:      session,
		basePath:     basePath,
		bom:          bom,
		layers:       layers,
		pollPause:    3 * time.Second,
		pollDuration: 600 * time.Second,
		pollInterval: 5 * time.Second,
	}
}

// AttachData will attach sbom, layers & policy to the handler.
func (h *Handler) AttachData(bom *Bom, layers []layers.Layer, buildStep, namespace, imageID string) {
	h.buildStep = buildStep
	h.namespace = namespace
	h.bom = bom
	h.layers = layers
	h.imageID = imageID
}

// HealthCheck will check the health of the service backend.
func (h Handler) HealthCheck() error {
	healthCheckPath := fmt.Sprintf(healthCheckTemplate, h.basePath)
	_, _, err := h.session.RequestData(http.MethodGet, healthCheckPath, nil)

	return err
}

// Scan will send payload to image scanning service and fetch the result back.
func (h *Handler) Scan(operationID string, opts Option) (*image.ScannedImage, error) {
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

		target, ok := h.bom.Packages.Source.Target.(bom.JSONImageSource)

		if !ok {
			errMsg := "Failed to get imageID"
			e := cberr.NewError(cberr.ScanFailedErr, errMsg, nil)
			logrus.Error(e.Error())

			return
		}

		h.imageID = target.ID

		if status, err := h.PutBomAndLayersToAnalysisAPI(operationID, opts); err != nil {
			errChan <- err
			return
		} else if status == FinishedStatus {
			// the result should be fetched directly from backend
			scannedImage, e := h.GetImageVulnerability(h.bom.ManifestDigest, "")
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

		scannedImage, err := h.GetResponseFromScanAPI(h.bom.ManifestDigest, operationID)
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

// PutBomAndLayersToAnalysisAPI will call the PUT API and upload sbom to image scanning service.
func (h Handler) PutBomAndLayersToAnalysisAPI(operationID string, opts Option) (Status, error) {
	versionInfo := version.GetCurrentVersion()

	payload := NewAnalysisPayload(&h.bom.Packages, h.layers, h.buildStep, h.namespace, opts.ForceScan, versionInfo.SyftVersion, versionInfo.Version)

	analysisPath := fmt.Sprintf(putSBOMTemplate, h.basePath, h.bom.ManifestDigest, operationID)
	if h.bom != nil && h.bom.FullTag != "" {
		statusURLWithQueries, _ := url.Parse(analysisPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		params.Add("image_id", h.imageID)
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
func (h Handler) GetResponseFromScanAPI(digest, operationID string) (*image.ScannedImage, error) {
	ticker := time.NewTicker(h.pollInterval)
	defer ticker.Stop()

	timeout := make(chan bool)
	statusResult := make(chan StatusResponse)

	go func() {
		time.Sleep(h.pollDuration)
		timeout <- true
	}()

	for {
		select {
		case <-timeout:
			return nil, cberr.NewError(cberr.TimeoutErr, "Time out during status polling", nil)
		case <-ticker.C:
			go func() {
				if status, err := h.GetImageAnalysisStatus(digest, operationID); err == nil {
					statusResult <- status
				} else {
					statusResult <- StatusResponse{OperationStatus: FailedStatus}
				}
			}()
		case result := <-statusResult:
			switch result.OperationStatus {
			case FinishedStatus:
				return h.GetImageVulnerability(digest, h.imageID)
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
func (h Handler) GetImageAnalysisStatus(digest, operationID string) (StatusResponse, error) {
	statusPath := fmt.Sprintf(getStatusTemplate, h.basePath, digest, operationID)
	if h.bom != nil && h.bom.FullTag != "" {
		statusURLWithQueries, _ := url.Parse(statusPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		statusURLWithQueries.RawQuery = params.Encode()
		statusPath = statusURLWithQueries.String()
	}

	var response StatusResponse
	statusCode, resp, err := h.session.RequestData(http.MethodGet, statusPath, nil)
	if err != nil {
		if cberr.ErrorCode(err) == cberr.HTTPUnsuccessfulResponseErr {
			errMsg := fmt.Sprintf("Failed to fetch status for image (%v)", digest)
			e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
			logrus.Errorln(e.Error())

			return response, e
		}

		return response, err
	}

	// the scanning message might still be in the queue, return QUEUED here
	if statusCode == http.StatusNoContent {
		response.OperationStatus = QueuedStatus
		return response, nil
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal scan status for image (%v)", digest)
		e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
		logrus.Errorln(e.Error())

		return response, e
	}

	return response, nil
}

// GetImageVulnerability will fetch the vulnerability result via image digest.
func (h Handler) GetImageVulnerability(digest string, imageID string) (*image.ScannedImage, error) {
	var vulnPath string

	if imageID != "" {
		vulnPath = fmt.Sprintf(getVulnByIdTemplate, h.basePath, imageID)
	} else {
		vulnPath = fmt.Sprintf(getVulnTemplate, h.basePath, digest)
	}

	if h.bom != nil && h.bom.FullTag != "" {
		vulnURLWithQueries, _ := url.Parse(vulnPath)
		params := url.Values{}
		params.Add("full_tag", h.bom.FullTag)
		vulnURLWithQueries.RawQuery = params.Encode()
		vulnPath = vulnURLWithQueries.String()
	}

	_, resp, err := h.session.RequestData(http.MethodGet, vulnPath, nil)
	if err != nil {
		var errMsg string
		if imageID != "" {
			errMsg = fmt.Sprintf("Failed to fetch scan report for image id (%v)", imageID)
		} else {
			errMsg = fmt.Sprintf("Failed to fetch scan report for image manifest(%v)", digest)
		}

		e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
		logrus.Errorln(e.Error())

		return nil, e
	}

	if len(resp) == 0 {
		var errMsg string
		if imageID != "" {
			errMsg = fmt.Sprintf("Empty respose for image id (%v)", imageID)
		} else {
			errMsg = fmt.Sprintf("Empty respose for image manifest (%v)", digest)
		}

		e := cberr.NewError(cberr.EmptyResponse, errMsg, err)
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

	// Unpacking packages from the sbom, if we want to unpack from backend, remove this line.
	if h.bom != nil {
		scannedImage.Packages = h.bom.Packages
	}

	return &scannedImage, nil
}

// GetImagesScanResultsFromBackendByImageID return scan image data if existed.
func (h Handler) GetImagesScanResultsFromBackendByImageID(imageId string) (*image.ScannedImage, error) {
	scannedImage, e := h.GetImageVulnerability("", imageId)
	return scannedImage, e
}
