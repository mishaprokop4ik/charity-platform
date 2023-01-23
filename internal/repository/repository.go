package repository

import models "Kurajj/internal/bussiness_models"

type Userer interface {
	CreateUser(user models.User) (uint, error)
	GetUser(id uint) (models.User, error)
	DeleteUser(id uint) error
	UpsertUser(newUser models.User) error
}

type Repository struct {
	user Userer
}

func New(user Userer) *Repository {
	return &Repository{
		user: user,
	}
}
