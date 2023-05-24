package service_test

import (
	"Kurajj/internal/models"
	service "Kurajj/internal/services"
	mock_service "Kurajj/internal/services/mocks"
	"bou.ke/monkey"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateProposalEvent(t *testing.T) {
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

func BenchmarkCreateProposalEvent(b *testing.B) {
	mockCtrl := gomock.NewController(b)
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

	for n := 0; n < b.N; n++ {
		repo.EXPECT().
			CreateProposalEvent(context.TODO(), proposalEvent)

		_, err := proposalEventService.CreateEvent(context.TODO(), proposalEvent)
		assert.NoError(b, err)
	}
}

func TestUpdateProposalEvent(t *testing.T) {
	// mocking
	event := models.ProposalEvent{
		ID:                    uint(1),
		Title:                 "Test",
		Description:           "Test",
		CreationDate:          time.Now(),
		AuthorID:              1,
		Status:                models.Active,
		MaxConcurrentRequests: 15,
		RemainingHelps:        5,
	}
	oldEvent := models.ProposalEvent{
		ID:                    uint(1),
		Title:                 "Test",
		Description:           "Test",
		CreationDate:          time.Now(),
		AuthorID:              1,
		Status:                models.Active,
		MaxConcurrentRequests: 10,
		RemainingHelps:        5,
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	proposalEventService := service.NewProposalEvent(repo)

	repo.EXPECT().GetEvent(context.TODO(), oldEvent.ID).Return(oldEvent, nil)
	newEvent := event
	newEvent.RemainingHelps = 10
	repo.EXPECT().UpdateEvent(context.TODO(), newEvent).Return(nil)

	err := proposalEventService.UpdateProposalEvent(context.TODO(), event)
	assert.NoError(t, err)
}

func TestGetProposalEventStatistics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	startDate := time.Now().AddDate(0, 0, -28)
	endData := time.Now()

	proposalEventService := service.NewProposalEvent(repo)

	repo.EXPECT().
		GetProposalEventStatistics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.Transaction{
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
		GetProposalEventStatistics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.Transaction{
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

	stats, err := proposalEventService.GetProposalEventStatistics(context.TODO(), 28, uint(1))
	assert.NoError(t, err)

	expectedStatistics := models.ProposalEventStatistics{
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

func TestAcceptWithCancelStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	repo.EXPECT().
		UpdateTransactionByID(context.TODO(), uint(1), map[string]any{
			"transaction_status": models.Canceled,
		})

	repo.EXPECT().
		GetTransactionByID(context.TODO(), uint(1)).Return(models.Transaction{
		ID:                1,
		CreatorID:         1,
		EventID:           2,
		EventType:         models.ProposalEventType,
		TransactionStatus: models.Canceled,
		ResponderStatus:   models.Canceled,
	}, nil)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	})

	repo.EXPECT().
		CreateNotification(context.TODO(), models.TransactionNotification{
			EventType:     models.ProposalEventType,
			EventID:       2,
			Action:        models.Updated,
			TransactionID: 1,
			NewStatus:     models.Canceled,
			IsRead:        false,
			CreationTime:  time.Now(),
			MemberID:      1,
		})

	proposalEventService := service.NewProposalEvent(repo)

	request := models.AcceptRequest{
		Accept:        false,
		TransactionID: 1,
		MemberID:      1,
	}

	err := proposalEventService.Accept(context.TODO(), request)
	assert.NoError(t, err)
}

func TestResponseWhenRequesterHasTransaction(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	})

	repo.EXPECT().
		GetEvent(context.TODO(), uint(1)).
		Return(models.ProposalEvent{
			ID:                    1,
			Title:                 "Test",
			Description:           "Test",
			CreationDate:          time.Now(),
			AuthorID:              1,
			Status:                models.Active,
			MaxConcurrentRequests: 10,
			RemainingHelps:        5,
		}, nil)

	proposalEventService := service.NewProposalEvent(repo)

	err := proposalEventService.Response(context.TODO(), 1, 1, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "event creator cannot response his/her own events")
}

func TestResponse(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	})

	repo.EXPECT().
		GetEvent(context.TODO(), uint(1)).
		Return(models.ProposalEvent{
			ID:                    1,
			Title:                 "Test",
			Description:           "Test",
			CreationDate:          time.Now(),
			AuthorID:              2,
			Status:                models.Active,
			MaxConcurrentRequests: 10,
			RemainingHelps:        5,
		}, nil)

	repo.EXPECT().
		UpdateRemainingHelps(context.TODO(), models.ID(1), false, 1)

	repo.EXPECT().
		CreateTransaction(context.TODO(), gomock.Any())

	repo.EXPECT().
		CreateNotification(context.TODO(), gomock.Any())

	proposalEventService := service.NewProposalEvent(repo)

	err := proposalEventService.Response(context.TODO(), 1, 1, "")
	assert.NoError(t, err)
}

func TestResponseWhenRequesterIsEventCreator(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	})

	repo.EXPECT().
		GetEvent(context.TODO(), uint(1)).
		Return(models.ProposalEvent{
			ID:                    1,
			Title:                 "Test",
			Description:           "Test",
			CreationDate:          time.Now(),
			AuthorID:              1,
			Status:                models.Active,
			MaxConcurrentRequests: 10,
			RemainingHelps:        5,
		}, nil)

	proposalEventService := service.NewProposalEvent(repo)

	err := proposalEventService.Response(context.TODO(), 1, 1, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "event creator cannot response his/her own events")
}

func TestResponseWhenRequesterAlreadyHasTransaction(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	})

	repo.EXPECT().
		GetEvent(context.TODO(), uint(1)).
		Return(models.ProposalEvent{
			ID:                    1,
			Title:                 "Test",
			Description:           "Test",
			CreationDate:          time.Now(),
			AuthorID:              1,
			Status:                models.Active,
			MaxConcurrentRequests: 10,
			RemainingHelps:        5,
			Transactions: []models.Transaction{
				{
					TransactionStatus: models.InProcess,
					ResponderStatus:   models.InProcess,
					CreatorID:         uint(2),
				},
			},
		}, nil)

	proposalEventService := service.NewProposalEvent(repo)

	err := proposalEventService.Response(context.TODO(), uint(1), uint(2), "")
	assert.Error(t, err)
	assert.EqualError(t, err, "user already has transaction in this event")
}

func TestAcceptWithAcceptStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := mock_service.NewMockRepositorier(mockCtrl)

	repo.EXPECT().
		UpdateTransactionByID(context.TODO(), uint(1), map[string]any{
			"transaction_status": models.Accepted,
		})

	repo.EXPECT().
		GetTransactionByID(context.TODO(), uint(1)).Return(models.Transaction{
		ID:                1,
		CreatorID:         1,
		EventID:           2,
		EventType:         models.ProposalEventType,
		TransactionStatus: models.Accepted,
		ResponderStatus:   models.Accepted,
	}, nil)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	})

	repo.EXPECT().
		CreateNotification(context.TODO(), models.TransactionNotification{
			EventType:     models.ProposalEventType,
			EventID:       2,
			Action:        models.Updated,
			TransactionID: 1,
			NewStatus:     models.Accepted,
			IsRead:        false,
			CreationTime:  time.Now(),
			MemberID:      1,
		})

	proposalEventService := service.NewProposalEvent(repo)

	request := models.AcceptRequest{
		Accept:        true,
		TransactionID: 1,
		MemberID:      1,
	}

	err := proposalEventService.Accept(context.TODO(), request)
	assert.NoError(t, err)
}

func generateZeroRequests(now time.Time, size int) []models.Request {
	requests := make([]models.Request, size)
	for i := 1; i <= size; i++ {
		currenntDateResponse := now.AddDate(0, 0, i)
		requests[i-1] = models.Request{
			Date: fmt.Sprintf("%s %d", currenntDateResponse.Month().String(), currenntDateResponse.Day()),
		}
	}

	return requests
}
