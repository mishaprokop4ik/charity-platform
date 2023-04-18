package models

import (
	"io"
	"time"
)

type HelpEvent struct {
	ID             uint      `gorm:"column:id"`
	Title          string    `gorm:"column:title"`
	Description    string    `gorm:"column:description"`
	Needs          []Need    `gorm:"gorm:"foreignkey:HelpEventID""`
	Tags           []Tag     `gorm:"-"`
	CreatedBy      uint      `gorm:"createdBy"`
	CreatedAt      time.Time `gorm:"createdAt"`
	CompletionTime time.Time `gorm:"completionTime"`
	ImagePath      string    `gorm:"column:image_path"`
	FileType       string    `gorm:"-"`
	File           io.Reader `gorm:"-"`
}
