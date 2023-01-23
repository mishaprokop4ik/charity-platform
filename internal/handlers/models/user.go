package models

import (
	httpHelper "Kurajj/pkg/http"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"net/http"
	"regexp"
)

type Email string

func (e *Email) IsEmail() (bool, error) {
	emailRegex, err := regexp.Compile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if err != nil {
		return false, fmt.Errorf("can not create email regex: %s", err)
	}

	return emailRegex.MatchString(string(*e)), nil
}

type Telephone string

func (t *Telephone) IsTelephone() bool {
	phoneNumber := *t
	if len(phoneNumber) == 14 {
		phoneNumber = phoneNumber[1:]
	}

	if len(phoneNumber) == 10 || len(phoneNumber) == 13 {
		return containsOnlyDigits(string(phoneNumber))
	}

	return false
}

func containsOnlyDigits(s string) bool {
	digits := []rune{
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
	}
	for i := range s {
		if !lo.Contains(digits, rune(s[i])) {
			return false
		}
	}

	return true
}

type Address struct {
	Region       string `json:"region,omitempty"`
	City         string `json:"city,omitempty"`
	District     string `json:"district,omitempty"`
	HomeLocation string `json:"homeLocation,omitempty"`
}

type NewUserInput struct {
	Email       Email     `json:"email,omitempty"`
	FirstName   string    `json:"firstName,omitempty"`
	SecondName  string    `json:"secondName,omitempty"`
	Telephone   Telephone `json:"telephone,omitempty"`
	CompanyName string    `json:"companyName,omitempty"`
	Password    string    `json:"password,omitempty"`
	Address     Address   `json:"address" json:"address"`
}

type GetUserInput struct {
	Email    Email  `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func UnmarshalNewUserInput(r *http.Request) (NewUserInput, error) {
	userBytes, err := httpHelper.GetBody(r)
	if err != nil {
		return NewUserInput{}, err
	}
	user := NewUserInput{}
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return NewUserInput{}, fmt.Errorf("cound not parse user input: %s", err)
	}
	return user, nil
}
