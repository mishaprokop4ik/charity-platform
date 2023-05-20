package models

import (
	"encoding/json"
	"io"
)

type FilePath struct {
	Path string `json:"path"`
}

func NewFilePath(from *io.ReadCloser) (FilePath, error) {
	path := FilePath{}
	err := json.NewDecoder(*from).Decode(&path)
	return path, err
}
