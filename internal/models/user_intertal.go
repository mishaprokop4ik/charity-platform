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
	Password         string
	Address          string
	IsDeleted        bool
	TelegramUsername string
	AvatarImagePath  string `gorm:"column:image_path"`
}

func (u User) getAddress() Address {
	fullAddress := strings.Split(u.Address, "|")
	if len(fullAddress) != DecodedAddressLength {
		panic(fmt.Sprintf("something went wrong. the size of address is incorrect. want %d; got: %d", DecodedAddressLength, len(fullAddress)))
	}

	return Address{
		Region:       fullAddress[0],
		City:         fullAddress[1],
		District:     fullAddress[2],
		HomeLocation: fullAddress[3],
	}
}

func (u User) GetUserFullResponse(token string) SignedInUser {
	fullName := strings.Split(u.FullName, " ")
	address := u.getAddress()
	return SignedInUser{
		ID: int(u.ID),
		//TODO add email validation
		Email:      Email(u.Email),
		FirstName:  fullName[0],
		SecondName: fullName[1],
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
