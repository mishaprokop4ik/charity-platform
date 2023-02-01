package repository

import (
	"Kurajj/internal/models"
	"context"
)

type Userer interface {
	CreateUser(ctx context.Context, user models.User) (uint, error)
	GetUserAuthentication(ctx context.Context, email, password string) (uint, error)
	GetEntity(ctx context.Context, email, password string, isAdmin, isDeleted bool) (models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, values map[string]any) error
	UpdateUserByEmail(ctx context.Context, email string, values map[string]any) error
}

type Repository struct {
	User  Userer
	Admin adminCRUDer
}

func New(dbConnector *Connector) *Repository {
	return &Repository{
		User:  NewUser(dbConnector),
		Admin: NewAdmin(dbConnector),
	}
}
