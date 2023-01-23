package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func SendErrorResponse(w http.ResponseWriter, status uint, message string) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, message, int(status))
}

func SendHTTPResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(data)
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
