package service

import (
	"Kurajj/internal/models"
	"context"
)

func NewComplaint(repo Repositorier) *Complaint {
	return &Complaint{repo: repo}
}

type Complaint struct {
	repo Repositorier
}

func (c *Complaint) Complain(ctx context.Context, complaint models.Complaint) (int, error) {
	return c.repo.Complain(ctx, complaint)
}

func (c *Complaint) GetAll(ctx context.Context) ([]models.ComplaintsResponse, error) {
	return c.repo.GetAll(ctx)
}

func (c *Complaint) BanUser(ctx context.Context, userID models.ID) error {
	return c.repo.BanUser(ctx, userID)
}

func (c *Complaint) BanEvent(ctx context.Context, eventID models.ID, eventType models.EventType) error {
	return c.repo.BanEvent(ctx, eventID, eventType)
}
