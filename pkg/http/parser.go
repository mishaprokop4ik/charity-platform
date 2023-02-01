package http

import (
	"fmt"
	"net/http"
)

func GetBody(r *http.Request) ([]byte, error) {
	data := []byte{}
	_, err := r.Body.Read(data)
	if err != nil {
		return []byte{}, fmt.Errorf("cound not get body from request: %s", err)
	}
	return data, nil
}
