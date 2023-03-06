package models

import (
	"errors"
)

var ErrNotFound = errors.New("no such entity")

type ErrResponse struct {
	Error string `json:"error"`
}
