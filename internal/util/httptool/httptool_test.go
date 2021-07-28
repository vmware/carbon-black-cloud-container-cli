/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package httptool_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/httptool"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
)

var (
	mockBasePath string
	session      *httptool.RequestSession
)

func TestMain(m *testing.M) {
	session = httptool.NewRequestSession("", "")
	mockServer := serverMock()
	mockBasePath = mockServer.URL

	code := m.Run()

	mockServer.Close()

	os.Exit(code)
}

func TestSendRequest_GetJSON(t *testing.T) {
	jsonURL := fmt.Sprintf("%s/json", mockBasePath)

	_, resp, err := session.RequestData(http.MethodGet, jsonURL, nil)
	if err != nil {
		t.Error(err)
	}

	var respMap map[string]interface{}

	if err := json.Unmarshal(resp, &respMap); err != nil {
		t.Errorf("failed to parse the resp: %v", err)
	}

	if respMap["result"] != "mock json" {
		t.Error("expected mock json response got", respMap)
	}
}

func TestSendRequest_PostJSON(t *testing.T) {
	jsonURL := fmt.Sprintf("%s/json", mockBasePath)

	_, resp, err := session.RequestData(http.MethodPost, jsonURL, nil)
	if err != nil {
		t.Error(err)
	}

	var respMap map[string]interface{}

	if err := json.Unmarshal(resp, &respMap); err != nil {
		t.Errorf("failed to parse the resp: %v", err)
	}

	if respMap["result"] != "mock json" {
		t.Error("expected mock json response got", respMap)
	}
}

func TestSendRequest_GetString(t *testing.T) {
	stringURL := fmt.Sprintf("%s/string", mockBasePath)

	_, resp, err := session.RequestData(http.MethodGet, stringURL, nil)
	if err != nil {
		t.Error(err)
	}

	if string(resp) != "mock string" {
		t.Error("expected mock string response got", resp)
	}
}

func TestSendRequest_Failures(t *testing.T) {
	failureMap := make(map[cberr.Code]string)
	failureMap[cberr.HTTPNotAllowedErr] = fmt.Sprintf("%s/unauthorized", mockBasePath)
	failureMap[cberr.HTTPNotFoundErr] = fmt.Sprintf("%s/not_found", mockBasePath)
	failureMap[cberr.HTTPUnsuccessfulResponseErr] = fmt.Sprintf("%s/internal_error", mockBasePath)

	for errorCode, url := range failureMap {
		_, _, err := session.RequestData(http.MethodGet, url, nil)
		if err == nil {
			t.Error("expect an error but not raised")
		}

		if cberr.ErrorCode(err) != errorCode {
			t.Errorf("expect error code: %v; actual error code: %v", errorCode, cberr.ErrorCode(err))
		}
	}
}

func TestTryReadErrorResponse(t *testing.T) {
	errorMsg := "{\"message\": \"test message\"}"
	errorBytes := []byte(errorMsg)

	msg, ok := httptool.TryReadErrorResponse(errorBytes)
	if ok == false || msg != "test message" {
		t.Errorf("wrong error message detected: %s", msg)
	}

	_, ok = httptool.TryReadErrorResponse([]byte(``))
	if ok == true {
		t.Error("should not parse the response")
	}
}

func serverMock() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{\"result\":\"mock json\"}"))
	})
	handler.HandleFunc("/string", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("mock string"))
	})
	handler.HandleFunc("/unauthorized", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	handler.HandleFunc("/not_found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	handler.HandleFunc("/internal_error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	return httptest.NewServer(handler)
}
