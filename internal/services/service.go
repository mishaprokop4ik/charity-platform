package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/repository"
)

type Service struct {
	Authentication Authenticator
	Admin          AdminCRUDer
}

func New(repo *repository.Repository, authConfig *config.AuthenticationConfig, emailConfig *config.Email) *Service {
	return &Service{
		Authentication: NewAuthentication(repo, authConfig, emailConfig),
		Admin:          NewAdmin(repo, authConfig, emailConfig),
	}
}
