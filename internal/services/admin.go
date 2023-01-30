package service

import (
	"Kurajj/internal/config"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

func NewAdmin(repo *repository.Repository, authConfig *config.AuthenticationConfig) *Admin {
	return &Admin{repo: repo, authConfig: authConfig}
}

type Admin struct {
	repo       *repository.Repository
	authConfig *config.AuthenticationConfig
}

func (a *Admin) CreateAdmin(ctx context.Context, admin models.User) (uint, error) {
	admin.Password = GeneratePasswordHash(admin.Password, a.authConfig.Salt)
	return a.repo.Admin.CreateAdmin(ctx, admin)
}

func (a *Admin) GetAdminByID(ctx context.Context, id uint) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Admin) UpdateAdmin(ctx context.Context, admin models.User) error {
	//TODO implement me
	panic("implement me")
}

func (a *Admin) DeleteAdmin(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}

func (a *Admin) GetAllAdmins(ctx context.Context) ([]models.User, error) {
	//TODO implement me
	panic("implement me")
}

type AdminCRUDer interface {
	CreateAdmin(ctx context.Context, admin models.User) (uint, error)
	GetAdminByID(ctx context.Context, id uint) (models.User, error)
	UpdateAdmin(ctx context.Context, admin models.User) error
	DeleteAdmin(ctx context.Context, id uint) error
	GetAllAdmins(ctx context.Context) ([]models.User, error)
}
