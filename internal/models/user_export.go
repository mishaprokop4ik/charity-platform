package models

import (
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

func (a Address) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", a.Region, a.City, a.District, a.HomeLocation)
}

type SignUpUser struct {
	Email       Email     `json:"email,omitempty"`
	FirstName   string    `json:"firstName,omitempty"`
	SecondName  string    `json:"secondName,omitempty"`
	Telephone   Telephone `json:"telephone,omitempty"`
	CompanyName string    `json:"companyName,omitempty"`
	Password    string    `json:"password,omitempty"`
	Address     Address   `json:"address" json:"address"`
}

type SignInUser struct {
	Email    Email  `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func UnmarshalSignUpUser(r *http.Request) (SignUpUser, error) {
	user := SignUpUser{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return SignUpUser{}, fmt.Errorf("cound not decode user: %s", err)
	}

	return user, nil
}

func UnmarshalSignInUser(r *http.Request) (SignInUser, error) {
	user := SignInUser{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return SignInUser{}, err
	}
	return user, nil
}

func (i SignUpUser) GetInternalUser() User {
	fullName := fmt.Sprintf("%s %s", i.FirstName, i.SecondName)
	return User{
		Email:       string(i.Email),
		FullName:    fullName,
		Telephone:   string(i.Telephone),
		CompanyName: i.CompanyName,
		Password:    i.Password,
		Address:     i.Address.String(),
	}
}

type UserCreationResponse struct {
	ID int `json:"ID,omitempty"`
}

type SignedInUser struct {
	ID          int       `json:"ID,omitempty"`
	Email       Email     `json:"email,omitempty"`
	FirstName   string    `json:"firstName,omitempty"`
	SecondName  string    `json:"secondName,omitempty"`
	Telephone   Telephone `json:"telephone,omitempty"`
	CompanyName string    `json:"companyName,omitempty"`
	Address     Address   `json:"address" json:"address"`
	Token       string    `json:"token,omitempty"`
}

func (s SignedInUser) Bytes() []byte {
	encoded, _ := json.Marshal(s)
	return encoded
}

func (r UserCreationResponse) Bytes() []byte {
	encoded, _ := json.Marshal(r)
	return encoded
}
