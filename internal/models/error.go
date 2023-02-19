package models

import (
	"errors"
)

var NotFoundError = errors.New("no such entity")

type ErrResponse struct {
	Error string `json:"error"`
}
