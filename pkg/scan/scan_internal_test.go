package scan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.bit9.local/octarine/cbctl/internal/util/httptool"
	"gitlab.bit9.local/octarine/cbctl/pkg/model/image"
)

var mockDigest = "digest"

func TestNewScanHandler(t *testing.T) {
	dummyHandler := NewScanHandler("defense-dev01.cbdtest.io/containers", "ABCD123", "FOO", "BAR", nil)

	if dummyHandler == nil {
		t.Errorf("failed to construct a scan handler")
		return
	}

	fakeBuildStep, fakeNamespace := "fake_build_step", "fake_namespace"
	dummyHandler.AttachSBOMBuildStepAndNamespace(nil, fakeBuildStep, fakeNamespace)

	if dummyHandler.buildStep != fakeBuildStep {
		t.Errorf("build step didn't attach correctly")
	}

	if dummyHandler.namespace != fakeNamespace {
		t.Errorf("namespace didn't attach correctly")
	}
}

func TestGetResponsePassNoForce(t *testing.T) {
	mockServer := passedServerMock()
	defer mockServer.Close()

	mockBasePath := mockServer.URL
	mockScanHandler := &Handler{
		bom:          &Bom{ManifestDigest: mockDigest},
		session:      httptool.NewRequestSession("", ""),
		basePath:     mockBasePath,
		pollDuration: 5 * time.Second,
		pollInterval: 2 * time.Millisecond,
	}

	resp, err := mockScanHandler.Scan(Option{ForceScan: false})
	if err != nil {
		t.Error(err)
	}

	if resp.ManifestDigest != mockDigest {
		t.Errorf("expected image digest: %v, got: %v", mockDigest, resp.ManifestDigest)
	}
}

func TestGetResponsePassWithForce(t *testing.T) {
	mockServer := passedServerMock()
	defer mockServer.Close()

	mockBasePath := mockServer.URL
	mockScanHandler := &Handler{
		bom:          &Bom{ManifestDigest: mockDigest, FullTag: mockDigest},
		session:      httptool.NewRequestSession("", ""),
		basePath:     mockBasePath,
		pollDuration: 5 * time.Second,
		pollInterval: 2 * time.Millisecond,
	}

	if err := mockScanHandler.HealthCheck(); err != nil {
		t.Error(err)
	}

	resp, err := mockScanHandler.Scan(Option{ForceScan: true})
	if err != nil {
		t.Error(err)
	}

	if resp.ManifestDigest != mockDigest {
		t.Errorf("expected image digest: %v, got: %v", mockDigest, resp.ManifestDigest)
	}
}

func TestGetResponse_Fail(t *testing.T) {
	mockServer := failedServerMock()
	defer mockServer.Close()

	mockBasePath := mockServer.URL
	mockScanHandler := &Handler{
		session:      httptool.NewRequestSession("", ""),
		basePath:     mockBasePath,
		pollDuration: 5 * time.Second,
		pollInterval: 2 * time.Millisecond,
	}

	resp, err := mockScanHandler.GetResponseFromScanAPI(mockDigest)
	if err == nil {
		t.Error("expect an error but got nil")
	}

	if resp != nil {
		t.Errorf("response should be a nil struct, got %v", resp)
	}
}

func passedServerMock() *httptest.Server {
	handler := http.NewServeMux()

	statusCallCnt := 0

	handler.HandleFunc(fmt.Sprintf(healthCheckTemplate, ""), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})

	handler.HandleFunc(fmt.Sprintf(getStatusTemplate, "", mockDigest), func(w http.ResponseWriter, r *http.Request) {
		statusCallCnt++

		result := "QUEUED"
		if statusCallCnt == 100 {
			result = "FINISHED"
		}

		_, _ = w.Write([]byte(result))
	})

	handler.HandleFunc(fmt.Sprintf(putSBOMTemplate, "", mockDigest), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})

	handler.HandleFunc(fmt.Sprintf(getVulnTemplate, "", mockDigest), func(w http.ResponseWriter, r *http.Request) {
		mockVuln := &image.ScannedImage{
			Identifier:      image.Identifier{ManifestDigest: "digest"},
			ImageMetadata:   image.Metadata{},
			Account:         "",
			ScanStatus:      "",
			Vulnerabilities: nil,
		}
		reqBodyBytes, _ := json.Marshal(mockVuln)
		_, _ = w.Write(reqBodyBytes)
	})

	return httptest.NewServer(handler)
}

func failedServerMock() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc(fmt.Sprintf(getStatusTemplate, "", mockDigest), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("QUEUED"))
	})

	handler.HandleFunc(fmt.Sprintf(getVulnTemplate, "", mockDigest), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{\"result\":\"mock vuln\"}"))
	})

	return httptest.NewServer(handler)
}
