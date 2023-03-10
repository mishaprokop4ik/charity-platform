package models

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"io"
	"regexp"
)

const ukrainePhoneNumberPrefix = "+380"

type Email string

type defaultFields struct {
	Email       Email     `json:"email"`
	FirstName   string    `json:"firstName"`
	SecondName  string    `json:"secondName"`
	Telephone   Telephone `json:"telephone"`
	CompanyName string    `json:"companyName"`
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
	ID           uint      `json:"-" gorm:"column:id" gorm:"primaryKey"`
	Region       string    `json:"region" gorm:"column:area"`
	City         string    `json:"city" gorm:"column:city"`
	District     string    `json:"district" gorm:"column:district"`
	HomeLocation string    `json:"homeLocation" gorm:"column:home"`
	Street       string    `json:"-" gorm:"column:street"`
	Country      string    `json:"-" gorm:"column:country"`
	EventType    EventType `json:"-" gorm:"column:event_type"`
	EventID      uint      `json:"-" gorm:"column:event_id"`
}

func (a Address) IsEmpty() bool {
	return a.Region == "" && a.City == "" && a.District == "" && a.HomeLocation == ""
}

func (a Address) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", a.Region, a.City, a.District, a.HomeLocation)
}

type SignUpUser struct {
	defaultFields
	Address  Address `json:"address"`
	Password string  `json:"password"`
}

// SignInEntity represents default sign in structure.
type SignInEntity struct {
	Email    Email  `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"-"`
}

func UnmarshalSignUpUser(r *io.ReadCloser) (SignUpUser, error) {
	user := SignUpUser{}
	if err := json.NewDecoder(*r).Decode(&user); err != nil {
		return SignUpUser{}, fmt.Errorf("cound not decode user: %s", err)
	}

	return user, nil
}

// UnmarshalSignInEntity gets an SignInEntity from http Request
func UnmarshalSignInEntity(r *io.ReadCloser) (SignInEntity, error) {
	e := SignInEntity{}
	err := json.NewDecoder(*r).Decode(&e)
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
	ID int `json:"id"`
}

type SignedInUser struct {
	ID           int                   `json:"id"`
	Email        Email                 `json:"email"`
	FirstName    string                `json:"firstName"`
	SecondName   string                `json:"secondName"`
	Telephone    Telephone             `json:"telephone"`
	CompanyName  string                `json:"companyName"`
	Address      Address               `json:"address"`
	AccessToken  string                `json:"token"`
	RefreshToken string                `json:"refreshToken"`
	SearchValues []SearchValueResponse `json:"searchValues"`
}

func (s SignedInUser) Bytes() []byte {
	encoded, _ := json.Marshal(s)
	return encoded
}

func (r CreationResponse) Bytes() []byte {
	encoded, _ := json.Marshal(r)
	return encoded
}

func UnmarshalCreateAdmin(r *io.ReadCloser) (AdminCreation, error) {
	admin := AdminCreation{}
	if err := json.NewDecoder(*r).Decode(&admin); err != nil {
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

type UserShortInfo struct {
	ID              uint      `json:"id"`
	Username        string    `json:"username"`
	ProfileImageURL string    `json:"profileImageURL"`
	PhoneNumber     Telephone `json:"phoneNumber"`
}

func (u UserShortInfo) TableName() string {
	return "members"
}
