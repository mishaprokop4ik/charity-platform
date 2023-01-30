package models

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

const DecodedAddressLength = 4

type Tabler interface {
	TableName() string
}

type User struct {
	gorm.Model
	ID               uint `gorm:"primaryKey"`
	Email            string
	FullName         string
	Telephone        string
	CompanyName      string
	IsAdmin          bool
	Password         string
	Address          string
	IsDeleted        bool
	TelegramUsername string
	AvatarImagePath  string `gorm:"column:image_path"`
}

func (u User) getAddress() (Address, error) {
	fullAddress := strings.Split(u.Address, "|")
	if len(fullAddress) != DecodedAddressLength {
		return Address{}, fmt.Errorf("something went wrong. the size of address is incorrect. want %d; got: %d", DecodedAddressLength, len(fullAddress))
	}

	return Address{
		Region:       fullAddress[0],
		City:         fullAddress[1],
		District:     fullAddress[2],
		HomeLocation: fullAddress[3],
	}, nil
}

func (u User) GetUserFullResponse(token string) SignedInUser {
	var (
		firstName  = ""
		secondName = ""
	)

	fullName := strings.Split(u.FullName, " ")
	if len(fullName) == 2 {
		firstName = fullName[0]
		secondName = fullName[1]
	}
	address, _ := u.getAddress()
	return SignedInUser{
		ID: int(u.ID),
		//TODO add email validation
		Email:      Email(u.Email),
		FirstName:  firstName,
		SecondName: secondName,
		//TODO add telephone validation
		Telephone:   Telephone(u.Telephone),
		CompanyName: u.CompanyName,
		Address:     address,
		Token:       token,
	}
}

func (User) TableName() string {
	return "members"
}
