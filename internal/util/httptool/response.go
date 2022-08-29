package httptool

import "encoding/json"

// TryReadErrorResponse tries to convert response body to an error message.
func TryReadErrorResponse(data []byte) (string, bool) {
	type ErrorResponse struct {
		Message string `json:"message"`
	}

	var errResponse ErrorResponse
	if err := json.Unmarshal(data, &errResponse); err != nil || errResponse.Message == "" {
		return "", false
	}

	return errResponse.Message, true
}
