package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
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
		Where("is_deleted = ?", false).
		First(&user).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("could not find with input email: %s; it may be besause the password is incorrect", email)
	}
	return user.ID, resp.Error
}

func (u *User) GetEntity(ctx context.Context, email, password string, isAdmin, isDeleted bool) (models.User, error) {
	user := models.User{}
	err := u.DBConnector.DB.
		WithContext(ctx).
		Where("email = ?", email).
		Where("password = ?", password).
		Where("is_admin = ?", isAdmin).
		Where("is_deleted = ?", isDeleted).
		First(&user).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, fmt.Errorf("could not found an entity ")
	}

	return user, err
}

func (u *User) DeleteUser(ctx context.Context, id uint) error {
	panic("")
}

func (u *User) UpsertUser(ctx context.Context, newUser models.User) error {
	panic("")
}
