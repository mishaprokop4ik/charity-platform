package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func SendErrorResponse(w http.ResponseWriter, statusCode uint, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(statusCode))
	resp := ErrorResponse{
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
	return nil
}
