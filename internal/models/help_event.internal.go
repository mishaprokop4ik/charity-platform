package models

import (
	"github.com/samber/lo"
	"io"
	"time"
)

type ID uint

type HelpEvent struct {
	ID                    uint          `gorm:"column:id"`
	Title                 string        `gorm:"column:title"`
	Description           string        `gorm:"column:description"`
	Needs                 []Need        `gorm:"gorm:foreignkey:HelpEventID"`
	Tags                  []Tag         `gorm:"-"`
	Status                string        `gorm:"column:status"`
	CreatedBy             uint          `gorm:"column:created_by"`
	CreatedAt             time.Time     `gorm:"column:created_at"`
	CompletionTime        time.Time     `gorm:"column:completion_time"`
	Comments              []Comment     `gorm:"-"`
	Transactions          []Transaction `gorm:"-"`
	TransactionNeeds      map[ID][]Need `gorm:"-"`
	User                  User          `gorm:"-"`
	ImagePath             string        `gorm:"column:image_path"`
	FileType              string        `gorm:"-"`
	File                  io.Reader     `gorm:"-"`
	CompletionPercentages float64       `gorm:"-"`
}

func (h *HelpEvent) CalculateCompletionPercentages() {
	allTransactions := float64(len(h.Needs))
	if allTransactions == 0 {
		return
	}
	finishedTransactions := float64(len(lo.Filter(h.Needs, func(need Need, index int) bool {
		return need.ReceivedTotal >= need.Amount
	})))
	h.CompletionPercentages = float64(finishedTransactions/allTransactions) * 100
}

func (h *HelpEvent) CalculateTransactionsCompletionPercentages() {
	for i, transaction := range h.Transactions {
		allNeedsCount := len(transaction.Needs)
		finishedNeeds := len(lo.Filter(transaction.Needs, func(need Need, index int) bool {
			return need.ReceivedTotal == need.Amount
		}))

		h.Transactions[i].CompletionPercentages = finishedNeeds / allNeedsCount
	}
}

func (h *HelpEvent) Response() HelpEventResponse {
	helpEventResponse := HelpEventResponse{
		ID:                    h.ID,
		Title:                 h.Title,
		Description:           h.Description,
		CreationDate:          h.CreatedAt,
		CompetitionDate:       h.CompletionTime,
		Status:                h.Status,
		ImageURL:              h.ImagePath,
		AuthorInfo:            h.User.ToShortInfo(),
		CompletionPercentages: h.CompletionPercentages,
	}
	comments := make([]CommentResponse, len(h.Comments))
	for i, comment := range h.Comments {
		updatedTime := ""
		if comment.UpdatedAt.Valid {
			updatedTime = comment.UpdatedAt.Time.String()
		}
		comments[i] = CommentResponse{
			ID:            comment.ID,
			Text:          comment.Text,
			CreationDate:  comment.CreationDate,
			IsUpdated:     comment.IsUpdated,
			UpdateTime:    updatedTime,
			UserShortInfo: comment.UserValues,
		}
	}
	helpEventResponse.Comments = comments
	tags := make([]TagResponse, len(h.Tags))
	for i, tag := range h.Tags {
		tags[i] = TagResponse{
			ID:     tag.ID,
			Title:  tag.Title,
			Values: tag.GetTagValuesResponse(),
		}
	}
	helpEventResponse.Tags = tags
	transactions := make([]HelpEventTransactionResponse, len(h.Transactions))
	for i := range h.Transactions {
		needs := make([]NeedResponse, len(h.TransactionNeeds[ID(h.Transactions[i].ID)]))
		for j := range h.TransactionNeeds[ID(h.Transactions[i].ID)] {
			needs[j] = NeedResponse{
				ID:            h.Needs[j].ID,
				Title:         h.Needs[j].Title,
				Amount:        h.Needs[j].Amount,
				ReceivedTotal: h.Needs[j].ReceivedTotal,
				Received:      h.Needs[j].Received,
				Unit:          h.Needs[j].Unit,
			}
		}
		finishData := ""
		if !h.Transactions[i].CompetitionDate.Time.IsZero() {
			finishData = h.Transactions[i].CompetitionDate.Time.String()
		}
		isApproved := h.Transactions[i].TransactionStatus == Completed

		allTransactions := float64(len(h.TransactionNeeds[ID(h.Transactions[i].ID)]))
		var completionPercentages float64
		if allTransactions != 0 {
			finishedTransactions := float64(len(lo.Filter(h.TransactionNeeds[ID(h.Transactions[i].ID)], func(need Need, index int) bool {
				return need.Received >= need.Amount
			})))
			completionPercentages = float64(finishedTransactions/allTransactions) * 100
		}

		transactions[i] = HelpEventTransactionResponse{
			TransactionID:         h.Transactions[i].ID,
			Needs:                 needs,
			CompetitionDate:       finishData,
			IsApproved:            isApproved,
			CompletionPercentages: completionPercentages,
		}
	}
	helpEventResponse.Transactions = transactions
	needs := make([]NeedResponse, len(h.Needs))
	for i := range h.Needs {
		needs[i] = NeedResponse{
			ID:            h.Needs[i].ID,
			Title:         h.Needs[i].Title,
			Amount:        h.Needs[i].Amount,
			ReceivedTotal: h.Needs[i].ReceivedTotal,
			Received:      h.Needs[i].Received,
			Unit:          h.Needs[i].Unit,
		}
	}
	helpEventResponse.Needs = needs
	return helpEventResponse
}

func (*HelpEvent) TableName() string {
	return "help_event"
}
