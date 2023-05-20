package repository

import (
	"Kurajj/internal/models"
	"context"
	"fmt"
)

func NewComplaint(db *Connector) *Complaint {
	return &Complaint{db}
}

type Complaint struct {
	*Connector
}

func (c *Complaint) Complain(ctx context.Context, complaint models.Complaint) (int, error) {
	err := c.DB.Create(complaint).WithContext(ctx).Error
	return int(complaint.ID), err
}

func (c *Complaint) GetAll(ctx context.Context) ([]models.ComplaintsResponse, error) {
	complaints := make([]models.Complaint, 0)
	err := c.DB.Find(&complaints).WithContext(ctx).Error
	if err != nil {
		return nil, err
	}
	complaintsResponse := make([]models.ComplaintsResponse, 0)
	for _, complaint := range complaints {
		if index, exists := containsEvent(complaintsResponse, complaint.EventID, complaint.EventType); exists {
			complaintsResponse[index].Complaints = append(complaintsResponse[index].Complaints, models.ComplaintResponse{
				Description:  complaint.Description,
				CreationDate: complaint.CreationDate,
			})
		} else {
			complaintResponse := models.ComplaintsResponse{
				EventID:   complaint.EventID,
				EventType: complaint.EventType,
				Complaints: []models.ComplaintResponse{
					{
						Description:  complaint.Description,
						CreationDate: complaint.CreationDate,
					},
				},
			}
			switch complaint.EventType {
			case models.HelpEventType:
				event, err := c.getHelpEventByID(ctx, complaint.EventID)
				if err != nil {
					return complaintsResponse, err
				}
				complaintResponse.CreatorEventID = int(event.CreatedBy)
				complaintResponse.CreationDate = event.CreatedAt
			case models.ProposalEventType:
				event, err := c.getProposalEvent(ctx, complaint.EventID)
				if err != nil {
					return complaintsResponse, err
				}
				complaintResponse.CreatorEventID = int(event.AuthorID)
				complaintResponse.CreationDate = event.CreationDate
			}

			complaintsResponse = append(complaintsResponse, complaintResponse)
		}
	}

	return complaintsResponse, nil
}

func (c *Complaint) BanUser(ctx context.Context, userID models.ID) error {
	tx := c.DB.Begin()
	err := tx.
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_blocked", true).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.
		Model(&models.HelpEvent{}).
		Where("created_by = ?", userID).
		Update("is_deleted", true).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.
		Model(&models.ProposalEvent{}).
		Where("created_by = ?", userID).
		Update("is_deleted", true).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (c *Complaint) BanEvent(ctx context.Context, eventID models.ID, eventType models.EventType) error {
	switch eventType {
	case models.HelpEventType:
		return c.banHelpEvent(ctx, eventID)
	case models.ProposalEventType:
		return c.banProposalEvent(ctx, eventID)
	default:
		return fmt.Errorf("no event type with %s name", eventType)
	}

	return nil
}

func (c *Complaint) banHelpEvent(ctx context.Context, eventID models.ID) error {
	return c.DB.
		Model(&models.HelpEvent{}).
		Update("is_banned", true).
		WithContext(ctx).
		Error
}

func (c *Complaint) banProposalEvent(ctx context.Context, eventID models.ID) error {
	return c.DB.
		Model(&models.ProposalEvent{}).
		Update("is_banned", true).
		WithContext(ctx).
		Error
}

func (c *Complaint) getHelpEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error) {
	event := models.HelpEvent{
		TransactionNeeds: map[models.ID][]models.Need{},
	}
	err := c.DB.First(&event, id).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}

	return event, err
}

func (c *Complaint) getProposalEvent(ctx context.Context, id models.ID) (models.ProposalEvent, error) {
	event := models.ProposalEvent{}
	err := c.DB.First(&event, id).WithContext(ctx).Error
	if err != nil {
		return models.ProposalEvent{}, err
	}

	return event, err
}

func containsEvent(complaints []models.ComplaintsResponse, eventID models.ID, eventType models.EventType) (int, bool) {
	for i, c := range complaints {
		if c.EventType == eventType && c.EventID == eventID {
			return i, true
		}
	}

	return -1, false
}
