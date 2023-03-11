package models

import (
	"encoding/json"
	"io"
	"time"
)

type MemberSession struct {
	RefreshToken string    `gorm:"column:refresh_token"`
	ExpiresAt    time.Time `gorm:"column:expires_at"`
	MemberID     uint      `gorm:"column:member_id" gorm:"primaryKey"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken"`
}

type TokensResponse struct {
	RefreshToken string `json:"refreshToken"`
	AccessToken  string `json:"accessToken"`
}

type Tokens struct {
	Access  string
	Refresh string
}

func (t TokensResponse) Bytes() []byte {
	bytes, _ := json.Marshal(t)
	return bytes
}

func ParseRefresh(from *io.ReadCloser) (RefreshTokenInput, error) {
	tokenInput := RefreshTokenInput{}
	err := json.NewDecoder(*from).Decode(&tokenInput)
	return tokenInput, err
}

func (MemberSession) TableName() string {
	return "member_session"
}
