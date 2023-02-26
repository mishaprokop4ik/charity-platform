package models

import (
	"database/sql"
	"time"
)

type CommentExport struct {
	ID           uint         `json:"id,omitempty"`
	Text         string       `json:"text,omitempty"`
	CreationDate time.Time    `json:"creationDate"`
	IsUpdated    bool         `json:"isUpdated,omitempty"`
	UpdateTime   sql.NullTime `json:"updateTime"`
	UserComment
}
