package service_test

import (
	"Kurajj/internal/models"
	service "Kurajj/internal/services"
	mock_service "Kurajj/internal/services/mocks"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetHelpEventStatistics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	startDate := time.Now().AddDate(0, 0, -28)
	endData := time.Now()

	helpEvent := service.NewHelpEvent(repo)

	repo.EXPECT().
		GetHelpEventStatistics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.Transaction{
		{
			ID:                2,
			CreatorID:         1,
			EventID:           1,
			CreationDate:      startDate,
			EventType:         models.ProposalEventType,
			TransactionStatus: models.InProcess,
			ResponderStatus:   models.Waiting,
		},
		{
			ID:                3,
			CreatorID:         1,
			EventID:           1,
			CreationDate:      startDate,
			EventType:         models.ProposalEventType,
			TransactionStatus: models.InProcess,
			ResponderStatus:   models.Waiting,
		},
	}, nil)
	repo.EXPECT().
		GetHelpEventStatistics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.Transaction{
		{
			ID:                1,
			CreatorID:         1,
			EventID:           1,
			CreationDate:      time.Now(),
			EventType:         models.ProposalEventType,
			TransactionStatus: models.InProcess,
			ResponderStatus:   models.Waiting,
		},
	}, nil)

	stats, err := helpEvent.GetHelpEventStatistics(context.TODO(), 28, uint(1))
	assert.NoError(t, err)

	expectedStatistics := models.HelpEventStatistics{
		DefaultStatistics: models.DefaultStatistics{
			Requests:          generateZeroRequests(startDate, 28),
			StartDate:         startDate,
			EndDate:           endData,
			TransactionsCount: 2,
			TransactionsCountCompareWithPreviousMonth: 100,
		},
	}
	assert.Equal(t, expectedStatistics.Requests, stats.Requests)
	assert.Equal(t, expectedStatistics.TransactionsCount, stats.TransactionsCount)
	assert.Equal(t, expectedStatistics.TransactionsCountCompareWithPreviousMonth, stats.TransactionsCountCompareWithPreviousMonth)
}
