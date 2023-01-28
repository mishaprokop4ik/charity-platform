package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error,omitempty"`
}

func SendErrorResponse(w http.ResponseWriter, statusCode uint, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(statusCode))
	resp := errorResponse{
		Error: message,
	}
	respEncoded, _ := json.Marshal(resp)
	_, _ = w.Write(respEncoded)
}

type Byter interface {
	Bytes() []byte
}

func SendHTTPResponse(w http.ResponseWriter, resp Byter) error {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(resp.Bytes())
	if err != nil {
		return fmt.Errorf("could not send response: %s", err)
	}
	// TODO add tries when w.Write did not send all bytes
	return nil
}

func createErrorResponse(message string) ([]byte, error) {
	response := struct {
		ErrorMessage string `json:"errorMessage"`
	}{
		ErrorMessage: message,
	}
	encodedResponse, err := json.Marshal(response)
	if err != nil {
		return []byte{}, fmt.Errorf("cound not encode error response: %s", err)
	}

	return encodedResponse, nil
}
