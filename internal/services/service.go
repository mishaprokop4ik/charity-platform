package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type Service struct {
	Authentication *Authentication
}

func New(repo *repository.Repository, config *config.AuthenticationConfig) *Service {
	return &Service{
		Authentication: NewAuthentication(repo, config),
	}
}

type Authenticator interface {
	SignUp(ctx context.Context, user models.User) (uint, error)
	SignIn(ctx context.Context, user models.User) (models.SignedInUser, error)
}
