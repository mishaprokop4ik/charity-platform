package service

import (
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
)

type UserSearcher interface {
	UpsertValues(ctx context.Context, userId uint, tags []models.MemberSearch) error
}

type UserSearch struct {
	repo *repository.Repository
}

func (u *UserSearch) UpsertValues(ctx context.Context, userId uint, tags []models.MemberSearch) error {
	return u.repo.UserSearchValue.UpsertTags(ctx, userId, tags)
}

func NewUserSearch(repo *repository.Repository) *UserSearch {
	return &UserSearch{
		repo: repo,
	}
}
