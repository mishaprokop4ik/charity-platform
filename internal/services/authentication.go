package service

import (
	"Kurajj/internal/handlers/models"
	"Kurajj/internal/repository"
	"context"
)

type Authentication struct {
	repo repository.Repository
}

func (a Authentication) SignUp(ctx context.Context, user models.NewUserInput) (uint, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authentication) SignIn(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
