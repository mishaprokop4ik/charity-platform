package repository

import (
	"Kurajj/internal/models"
	"context"
)

type Userer interface {
	CreateUser(ctx context.Context, user models.User) (uint, error)
	GetUserAuthentication(ctx context.Context, email, password string) (uint, error)
	GetUser(ctx context.Context, id uint) (models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, newUser models.User) error
}

type Repository struct {
	User Userer
}

func New(dbConnector *Connector) *Repository {
	return &Repository{
		User: NewUser(dbConnector),
	}
}
