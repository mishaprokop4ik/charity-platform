package service

import (
	models "Kurajj/internal/bussiness_models"
	"context"
)

type Service struct {
	Authentication Authentication
}

func New(auth Authentication) *Service {
	return &Service{
		Authentication: auth,
	}
}

type Authenticator interface {
	SignUp(ctx context.Context, user models.User) (uint, error)
	SignIn(ctx context.Context, user models.User) (*models.User, error)
}
