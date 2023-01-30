package repository

import (
	"Kurajj/internal/models"
	"context"
)

type Admin struct {
	DBConnector *Connector
}

func (a *Admin) CreateAdmin(ctx context.Context, admin models.User) (uint, error) {
	err := a.DBConnector.DB.
		Create(&admin).
		WithContext(ctx).
		Error

	return admin.ID, err
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

type adminCRUDer interface {
	CreateAdmin(ctx context.Context, admin models.User) (uint, error)
	GetAdminByID(ctx context.Context, id uint) (models.User, error)
	UpdateAdmin(ctx context.Context, admin models.User) error
	DeleteAdmin(ctx context.Context, id uint) error
	GetAllAdmins(ctx context.Context) ([]models.User, error)
}

func NewAdmin(DBConnector *Connector) *Admin {
	return &Admin{DBConnector: DBConnector}
}
