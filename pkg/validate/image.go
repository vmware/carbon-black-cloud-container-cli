package validate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/httptool"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/image"
)

const (
	validateImageTemplate       = "%s/guardrails/validator/%s/image/%s"
	updateImageMetadataTemplate = "%s/management/images/%s/metadata"
)

// ImageHandler has all the fields for sending request to validator service.
type ImageHandler struct {
	session          *httptool.RequestSession
	imageID          image.Identifier
	basePath         string
	scanningBasePath string
	buildStep        string
	imageDigest      string
	namespace        string
}

// NewImageValidateHandler will create a handler for validate cmd.
func NewImageValidateHandler(saasTmpl, orgKey, apiID, apiKey, buildStep, namespace, imageDigest string) *ImageHandler {
	saasTmpl = strings.Trim(saasTmpl, "/")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/orgs")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1beta")

	// for guardrails service, we need to remove beta in the version if added
	// should remove when image-scanning-service not use beta version
	basePath := fmt.Sprintf("%s/v1/orgs/%s", strings.Trim(saasTmpl, "/"), orgKey)
	scanningBasePath := fmt.Sprintf("%s/v1beta/orgs/%s", strings.Trim(saasTmpl, "/"), orgKey)
	session := httptool.NewRequestSession(apiID, apiKey)

	return newValidateHandler(session, basePath, scanningBasePath, buildStep, namespace, imageDigest)
}

func newValidateHandler(
	session *httptool.RequestSession,
	basePath, scanningBasePath, buildStep, namespace, imageDigest string,
) *ImageHandler {
	return &ImageHandler{
		session:          session,
		basePath:         basePath,
		scanningBasePath: scanningBasePath,
		buildStep:        buildStep,
		namespace:        namespace,
		imageDigest:      imageDigest,
	}
}

// AttachImageID will attach image digest and identifier to the handler.
func (h *ImageHandler) AttachImageID(imageDigest string, imageID image.Identifier) {
	h.imageID = imageID
	h.imageDigest = imageDigest
}

// Validate will will validate the image using the validate API.
func (h ImageHandler) Validate() ([]image.PolicyViolation, error) {
	violations, err := h.getImageViolations()
	if err != nil {
		return nil, err
	}

	updateErr := h.updateImageMetadata()
	if updateErr != nil {
		logrus.Warnf("failed to update image metadata: %v", updateErr)
	}

	return violations, nil
}

func (h ImageHandler) updateImageMetadata() error {
	updateMetadataURL, err := url.Parse(fmt.Sprintf(updateImageMetadataTemplate, h.scanningBasePath, h.imageDigest))
	if err != nil {
		return fmt.Errorf("unexpected error, failed to parse update metadata url: %v", err)
	}

	type UpdateScanMetadataRequest struct {
		image.Identifier `json:",inline"`
		BuildStep        string `json:"build_step"`
		Namespace        string `json:"namespace"`
	}

	payload := UpdateScanMetadataRequest{
		Identifier: h.imageID,
		BuildStep:  h.buildStep,
		Namespace:  h.namespace,
	}

	_, _, updateErr := h.session.RequestData(http.MethodPut, updateMetadataURL.String(), payload)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (h ImageHandler) getImageViolations() ([]image.PolicyViolation, error) {
	if err := CheckValidBuildStep(h.buildStep); err != nil {
		return nil, err
	}

	validateURL, err := url.Parse(fmt.Sprintf(validateImageTemplate, h.basePath, h.buildStep, h.imageDigest))
	if err != nil {
		return nil, fmt.Errorf("unexpected error, failed to parse validate url: %v", err)
	}

	if h.namespace != "" {
		params := url.Values{}
		params.Add("namespace", h.namespace)
		validateURL.RawQuery = params.Encode()
	}

	_, resp, err := h.session.RequestData(http.MethodGet, validateURL.String(), nil)
	if err != nil {
		errMsg := "Failed to get violations from backend"
		e := cberr.NewError(cberr.ScanFailedErr, errMsg, err)
		logrus.Error(e.Error())

		return nil, err
	}

	type ImageValidatorResponse struct {
		PolicyViolations []image.PolicyViolation `json:"policy_violations"`
	}

	var result ImageValidatorResponse

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.PolicyViolations, nil
}
