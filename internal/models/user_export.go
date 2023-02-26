package models

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"net/http"
	"regexp"
)

const ukrainePhoneNumberPrefix = "+380"

type Email string

type defaultFields struct {
	Email       Email     `json:"email,omitempty"`
	FirstName   string    `json:"firstName,omitempty"`
	SecondName  string    `json:"secondName,omitempty"`
	Telephone   Telephone `json:"telephone,omitempty"`
	CompanyName string    `json:"companyName,omitempty"`
}

func (e Email) Validate() (bool, error) {
	emailRegex, err := regexp.Compile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if err != nil {
		return false, fmt.Errorf("can not create email regex: %s", err)
	}

	return emailRegex.MatchString(string(e)), nil
}

type Telephone string

func (t Telephone) GetDefaultTelephoneNumber() Telephone {
	phoneNumber := t
	if len(t) == 10 {
		phoneNumber = ukrainePhoneNumberPrefix + phoneNumber
	}

	return phoneNumber
}

func (t Telephone) Validate() bool {
	phoneNumber := t
	if len(phoneNumber) == 13 {
		phoneNumber = phoneNumber[1:]
	}
	if len(phoneNumber) == 10 || len(phoneNumber) == 12 {
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
	defaultFields
	Address  Address `json:"address"`
	Password string  `json:"password,omitempty"`
}

// SignInEntity represents default sign in structure.
type SignInEntity struct {
	Email    Email  `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	IsAdmin  bool   `json:"-"`
}

func UnmarshalSignUpUser(r *http.Request) (SignUpUser, error) {
	user := SignUpUser{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return SignUpUser{}, fmt.Errorf("cound not decode user: %s", err)
	}

	return user, nil
}

// UnmarshalSignInEntity gets an SignInEntity from http Request
func UnmarshalSignInEntity(r *http.Request) (SignInEntity, error) {
	e := SignInEntity{}
	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		return SignInEntity{}, err
	}
	return e, nil
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

type CreationResponse struct {
	ID int `json:"id,omitempty"`
}

type SignedInUser struct {
	ID          int       `json:"id,omitempty"`
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

func (r CreationResponse) Bytes() []byte {
	encoded, _ := json.Marshal(r)
	return encoded
}

func UnmarshalCreateAdmin(r *http.Request) (AdminCreation, error) {
	admin := AdminCreation{}
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		return AdminCreation{}, fmt.Errorf("cound not decode user: %s", err)
	}

	// TODO add validation for company name
	admin.CompanyName = "nure"

	return admin, nil
}

type AdminCreation struct {
	defaultFields
	IsAdmin bool `json:"-"`
}

func (a AdminCreation) CreateUser() User {
	fullName := fmt.Sprintf("%s %s", a.FirstName, a.SecondName)
	return User{
		Email:       string(a.Email),
		FullName:    fullName,
		Telephone:   string(a.Telephone),
		CompanyName: a.CompanyName,
		IsAdmin:     true,
		IsActivated: true,
	}
}

type UserComment struct {
	AuthorID        uint   `json:"authorId,omitempty"`
	Username        string `json:"username,omitempty"`
	ProfileImageURL string `json:"profileImageURL,omitempty"`
}
