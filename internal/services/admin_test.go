package service_test

import (
	"Kurajj/internal/models"
	service "Kurajj/internal/services"
	mock_service "Kurajj/internal/services/mocks"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAdmin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	proposalEventService := service.NewProposalEvent(repo)

	proposalEvent := models.ProposalEvent{
		Title:                 "Title",
		Description:           "Description",
		AuthorID:              1,
		Status:                models.Active,
		MaxConcurrentRequests: 5,
		Location: models.Address{
			Region:       "Kharkiv Oblast",
			City:         "Krarkiv",
			District:     "NovaBavaria",
			HomeLocation: "21",
			Street:       "Test",
			Country:      "Ukraine",
		},
	}

	repo.EXPECT().
		CreateProposalEvent(context.TODO(), proposalEvent)

	_, err := proposalEventService.CreateEvent(context.TODO(), proposalEvent)
	assert.NoError(t, err)
}
