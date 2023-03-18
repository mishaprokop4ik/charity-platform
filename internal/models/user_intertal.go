package models

import (
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

const DecodedAddressLength = 4

type Tabler interface {
	TableName() string
}

type User struct {
	gorm.Model
	ID               uint           `gorm:"primaryKey"`
	Email            string         `gorm:"column:email"`
	FullName         string         `gorm:"column:full_name"`
	Telephone        string         `gorm:"column:telephone"`
	CompanyName      string         `gorm:"column:company_name"`
	IsAdmin          bool           `gorm:"column:is_admin"`
	Password         string         `gorm:"column:password"`
	Address          string         `gorm:"column:address"`
	IsDeleted        bool           `gorm:"column:is_deleted"`
	IsActivated      bool           `gorm:"column:is_activated"`
	TelegramUsername string         `gorm:"column:telegram_username"`
	AvatarImagePath  string         `gorm:"column:image_path"`
	UserSearchValues []MemberSearch `gorm:"-"`
	Token            string         `json:"token"`
	RefreshToken     string         `json:"refreshToken"`
}

func (u User) ToShortInfo() UserShortInfo {
	return UserShortInfo{
		ID:              u.ID,
		Username:        u.FullName,
		ProfileImageURL: u.AvatarImagePath,
		PhoneNumber:     Telephone(u.Telephone),
	}
}

func (u User) getAddress() (Address, error) {
	fullAddress := strings.Split(u.Address, "|")
	if len(fullAddress) != DecodedAddressLength {
		return Address{},
			fmt.Errorf("something went wrong. the size of address is incorrect. want %d; got: %d",
				DecodedAddressLength, len(fullAddress))
	}

	return Address{
		Region:       fullAddress[0],
		City:         fullAddress[1],
		District:     fullAddress[2],
		HomeLocation: fullAddress[3],
	}, nil
}

func (u User) GetValuesToUpdate() map[string]any {
	getUserTag := func(f reflect.StructField, tagName string) string {
		tag := strings.Split(f.Tag.Get(tagName), ":")
		if len(tag) != 2 {
			return ""
		}
		return tag[1]
	}
	updateValues := make(map[string]any)

	user := reflect.TypeOf(u)
	userFields := reflect.ValueOf(u)
	userFieldsCount := user.NumField()
	for i := 0; i < userFieldsCount; i++ {
		field := user.Field(i)
		value := userFields.Field(i).Interface()
		if !userFields.Field(i).IsZero() {
			updateValues[getUserTag(field, "gorm")] = value
		}
	}

	return updateValues
}

func (u User) GetUserFullResponse(tokens Tokens) SignedInUser {
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
	searchValues := make([]SearchValueResponse, len(u.UserSearchValues))
	for i, searchValue := range u.UserSearchValues {
		searchValues[i] = searchValue.Response()
	}
	return SignedInUser{
		ID: int(u.ID),
		//TODO add email validation
		Email:      Email(u.Email),
		FirstName:  firstName,
		SecondName: secondName,
		//TODO add telephone validation
		Telephone:    Telephone(u.Telephone),
		CompanyName:  u.CompanyName,
		Address:      address,
		Avatar:       u.AvatarImagePath,
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
		SearchValues: searchValues,
	}
}

func (User) TableName() string {
	return "members"
}
