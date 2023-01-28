package repository

import (
	"Kurajj/internal/models"
	"context"
)

type User struct {
	DBConnector *Connector
}

func NewUser(DBConnector *Connector) *User {
	return &User{DBConnector: DBConnector}
}

func (u *User) CreateUser(ctx context.Context, user models.User) (uint, error) {
	// TODO add saving avatar image in S3
	err := u.DBConnector.DB.
		Create(&user).
		WithContext(ctx).Error
	return user.ID, err
}

func (u *User) GetUserAuthentication(ctx context.Context, email, password string) (uint, error) {
	user := models.User{}
	resp := u.DBConnector.DB.
		Where("password = ?", password).
		Where("email = ?", email).
		First(&user).
		WithContext(ctx)
	return user.ID, resp.Error
}

func (u *User) GetUser(ctx context.Context, id uint) (models.User, error) {
	user := models.User{}
	resp := u.DBConnector.DB.
		Where("id = ?", id).
		First(&user).
		WithContext(ctx)
	return user, resp.Error
}

func (u *User) DeleteUser(ctx context.Context, id uint) error {
	panic("")
}

func (u *User) UpsertUser(ctx context.Context, newUser models.User) error {
	panic("")
}
