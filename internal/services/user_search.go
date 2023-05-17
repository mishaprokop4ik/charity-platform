package service

import (
	"Kurajj/internal/models"
	"context"
)

func NewUserSearch(repo Repositorier) *UserSearch {
	return &UserSearch{
		repo: repo,
	}
}

type UserSearch struct {
	repo Repositorier
}

func (u *UserSearch) UpsertValues(ctx context.Context, userId uint, tags []models.MemberSearch) error {
	return u.repo.UpsertUserTags(ctx, userId, tags)
}
