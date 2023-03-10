package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Userer interface {
	CreateUser(ctx context.Context, user models.User) (uint, error)
	GetUserAuthentication(ctx context.Context, email, password string) (uint, error)
	GetUserInfo(ctx context.Context, id uint) (models.User, error)
	GetEntity(ctx context.Context, email, password string, isAdmin, isDeleted bool) (models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, values map[string]any) error
	UpdateUserByEmail(ctx context.Context, email string, values map[string]any) error
	IsEmailTaken(ctx context.Context, email string) (bool, error)
}

type User struct {
	DBConnector *Connector
}

func (u *User) GetUserInfo(ctx context.Context, id uint) (models.User, error) {
	user := models.User{}
	err := u.DBConnector.DB.
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		Where("is_activated = ?", true).
		First(&user).
		WithContext(ctx).
		Error
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""

	return user, nil
}

func NewUser(DBConnector *Connector) *User {
	return &User{DBConnector: DBConnector}
}

func (u *User) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := u.DBConnector.DB.Model(&models.User{}).
		Select("count(*) > 0").
		Where("email = ?", email).
		Find(&exists).
		Error

	return exists, err
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
		Where("is_activated = ?", true).
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

func (u *User) UpsertUser(ctx context.Context, values map[string]any) error {
	//return := u.DBConnector.DB.Omit()
	panic("")
}

func (u *User) UpdateUserByEmail(ctx context.Context, email string, values map[string]any) error {
	return u.DBConnector.DB.
		Model(&models.User{}).
		Select(lo.Keys(values)).
		Where("email = ?", email).
		Updates(values).
		Error
}
