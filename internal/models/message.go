package models

import (
	"encoding/json"
	"io"
)

type ConfirmMessage struct {
	Text string
	To   string
}

type UserConfirm struct {
	UserID      int     `json:"userID,omitempty"`
	ConfirmCode []int64 `json:"confirmCode,omitempty"`
}

func UnmarshalUserConfirm(r *io.ReadCloser) (UserConfirm, error) {
	e := UserConfirm{}
	err := json.NewDecoder(*r).Decode(&e)
	if err != nil {
		return UserConfirm{}, err
	}
	return e, nil
}
