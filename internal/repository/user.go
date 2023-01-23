package repository

import models "Kurajj/internal/bussiness_models"

type User struct {
	DBConnector Connector
}

func (u *User) CreateUser(user models.User) (uint, error) {
	panic("")
}

func (u *User) GetUser(id uint) (models.User, error) {
	panic("")
}

func (u *User) DeleteUser(id uint) error {
	panic("")
}

func (u *User) UpsertUser(newUser models.User) error {
	panic("")
}
