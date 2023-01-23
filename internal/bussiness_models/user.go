package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email       string
	FirstName   string
	SecondName  string
	Telephone   string
	CompanyName string
	Password    string
	Address     string
}
