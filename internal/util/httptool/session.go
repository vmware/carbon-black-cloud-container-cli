/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package httptool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
)

// CarbonBlackAccessInfo contains the api id and api key for Carbon Black.
type CarbonBlackAccessInfo struct {
	apiID  string
	apiKey string
}

// GetToken generates auth token from api key & id,
// A normal Carbon Black Access Token will be in the format of '<api-key>/<api-id>'.
func (c CarbonBlackAccessInfo) GetToken() string {
	return fmt.Sprintf("%s/%s", c.apiKey, c.apiID)
}

// RequestSession is the session used for sending request.
type RequestSession struct {
	client       http.Client
	cbAccessInfo *CarbonBlackAccessInfo
}

// NewRequestSession creates a new request session.
func NewRequestSession(apiID, apiKey string) *RequestSession {
	cbAccessInfo := &CarbonBlackAccessInfo{
		apiID:  apiID,
		apiKey: apiKey,
	}

	return &RequestSession{
		client: http.Client{
			Timeout: 300 * time.Second,
		},
		cbAccessInfo: cbAccessInfo,
	}
}

// RequestData will request data via method to url with payload.
func (r RequestSession) RequestData(method, url string, payload interface{}) (int, []byte, error) {
	var msg string

	req, err := r.generateRequest(method, url, payload)
	if err != nil {
		return 0, nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil || resp == nil {
		msg = "Failed to establish connection to Carbon Black Cloud, please check your config"
		e := cberr.NewError(cberr.HTTPConnectionErr, msg, err)
		logrus.Errorf("Failed to retrieve response: [%s] %s: %v", method, url, e)

		return 0, nil, e
	}

	defer func() {
		if e := resp.Body.Close(); e != nil {
			logrus.Errorf("Error when closing response: %v", e)
		}
	}()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg = fmt.Sprintf("Failed to read response body: [%s] %s", method, url)
		e := cberr.NewError(cberr.HTTPConnectionErr, msg, err)
		logrus.Errorln(e)

		return resp.StatusCode, nil, e
	}

	switch c := resp.StatusCode; {
	case c == http.StatusUnauthorized || c == http.StatusForbidden:
		msg = "The requested resource is restricted, please check API access"
		return c, nil, cberr.NewError(cberr.HTTPNotAllowedErr, msg, nil)
	case c == http.StatusNotFound:
		msg = "The requested resource not found"
		return c, nil, cberr.NewError(cberr.HTTPNotFoundErr, msg, nil)
	case c >= http.StatusMultipleChoices:
		msg = fmt.Sprintf("Unsuccessful %d response", c)
		e := cberr.NewError(cberr.HTTPUnsuccessfulResponseErr, msg, nil)
		logrus.Errorf("Unsuccessful response received: %v; error: %v", string(respBody), e.Error())

		return c, respBody, e
	default:
		return c, respBody, nil
	}
}

func (r RequestSession) generateRequest(method, url string, payload interface{}) (*http.Request, error) {
	var requester *http.Request

	switch method {
	case http.MethodGet, http.MethodDelete:
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return &http.Request{}, err
		}

		requester = req
	case http.MethodPost, http.MethodPatch, http.MethodPut:
		buffer := new(bytes.Buffer)

		if err := json.NewEncoder(buffer).Encode(payload); err != nil {
			return &http.Request{}, err
		}

		req, err := http.NewRequest(method, url, buffer)
		if err != nil {
			return &http.Request{}, err
		}

		requester = req
		requester.Header.Set("Content-Type", "application/json")
	default:
		return nil, fmt.Errorf("request session is not supporting method: %v", method)
	}

	accessToken := r.cbAccessInfo.GetToken()
	requester.Header.Set("X-Auth-Token", accessToken)

	return requester, nil
}
