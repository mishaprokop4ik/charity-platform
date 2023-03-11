package repository

import (
	"Kurajj/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Userer interface {
	CreateUser(ctx context.Context, user models.User) (uint, error)
	GetUserAuthentication(ctx context.Context, email, password string) (models.User, error)
	GetUserInfo(ctx context.Context, id uint) (models.User, error)
	GetEntity(ctx context.Context, email, password string, isAdmin, isDeleted bool) (models.User, error)
	SetSession(ctx context.Context, userID uint, session models.MemberSession) error
	GetByRefreshToken(ctx context.Context, token string) (models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, values map[string]any) error
	UpdateUserByEmail(ctx context.Context, email string, values map[string]any) error
	IsEmailTaken(ctx context.Context, email string) (bool, error)
}

type User struct {
	DBConnector *Connector
}

func (u *User) SetSession(ctx context.Context, userID uint, session models.MemberSession) error {
	session.MemberID = userID
	err := u.DBConnector.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "member_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"refresh_token", "expires_at"}),
	}).Create(&session).
		WithContext(ctx).
		Error

	return err
}

func (u *User) GetByRefreshToken(ctx context.Context, token string) (models.User, error) {
	session := models.MemberSession{}
	err := u.DBConnector.DB.Where("refresh_token = ?", token).First(&session).WithContext(ctx).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, fmt.Errorf("the token may be expired")
	} else if err != nil {
		return models.User{}, err
	}
	fmt.Println("userId", session.MemberID)
	member := models.User{}
	err = u.DBConnector.DB.First(&member, session.MemberID).WithContext(ctx).Error
	fmt.Println("member", member)
	return member, err
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

func (u *User) GetUserAuthentication(ctx context.Context, email, password string) (models.User, error) {
	user := models.User{}
	resp := u.DBConnector.DB.
		Where("password = ?", password).
		Where("email = ?", email).
		Where("is_deleted = ?", false).
		Where("is_activated = ?", true).
		First(&user).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return user, fmt.Errorf("could not find with input email: %s; it may be besause the password is incorrect", email)
	}

	return user, resp.Error
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
		Where("email = ?", email).
		Select(lo.Keys(values)).
		Updates(values).
		Error
}
