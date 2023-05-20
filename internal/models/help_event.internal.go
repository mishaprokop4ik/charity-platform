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
	Status                EventStatus   `gorm:"column:status"`
	CreatedBy             uint          `gorm:"column:created_by"`
	CreatedAt             time.Time     `gorm:"column:creation_date"`
	CompletionTime        time.Time     `gorm:"column:completion_time"`
	Banned                bool          `gorm:"column:is_banned"`
	Comments              []Comment     `gorm:"-"`
	Transactions          []Transaction `gorm:"-"`
	Location              Address       `gorm:"-"`
	TransactionNeeds      map[ID][]Need `gorm:"-"`
	User                  User          `gorm:"-"`
	ImagePath             string        `gorm:"column:image_path"`
	FileType              string        `gorm:"-"`
	File                  io.Reader     `gorm:"-"`
	CompletionPercentages float64       `gorm:"-"`
}

func (h *HelpEvent) CalculateCompletionPercentages() {
	h.CalculateTransactionsCompletionPercentages()
	if len(h.Needs) == 0 {
		return
	}
	eventCompletionPercentages := h.calculateCompletionPercentagesForNeeds(h.Needs)

	h.CompletionPercentages = eventCompletionPercentages / float64(len(h.Needs))
}

func (h *HelpEvent) CalculateTransactionsCompletionPercentages() {
	needsCompetitionPercentages := 0
	for i, transaction := range h.Transactions {
		for _, need := range transaction.Needs {
			if need.Amount == 0 {
				continue
			}
			needsCompetitionPercentages += int(need.Received/need.Amount) * 100
		}

		h.Transactions[i].CompletionPercentages = needsCompetitionPercentages
	}
}

func (h *HelpEvent) calculateCompletionPercentagesForNeeds(needs []Need) float64 {
	needsCompetitionPercentages := float64(0)
	for _, need := range needs {
		if need.Amount == 0 {
			continue
		}
		needsCompetitionPercentages += float64(need.ReceivedTotal/need.Amount) * 100
	}
	return needsCompetitionPercentages
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
	homeLocation := ""
	if h.Location.Street != "" {
		homeLocation = h.Location.Street
	}
	if h.Location.HomeLocation != "" && h.Location.Street != "" {
		homeLocation = homeLocation + " " + h.Location.HomeLocation
	} else if h.Location.HomeLocation != "" && h.Location.Street == "" {
		homeLocation = h.Location.HomeLocation
	}
	tags = append(tags, TagResponse{
		ID:    h.Location.ID,
		Title: "location",
		Values: []string{
			h.Location.Region,
			h.Location.City,
			h.Location.District,
			homeLocation,
		},
	})
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
		transactionNeeds := h.TransactionNeeds[ID(h.Transactions[i].ID)]
		needs := make([]NeedResponse, len(transactionNeeds))
		for j := range transactionNeeds {
			needs[j] = NeedResponse{
				ID:            transactionNeeds[j].ID,
				Title:         transactionNeeds[j].Title,
				Amount:        transactionNeeds[j].Amount,
				ReceivedTotal: transactionNeeds[j].ReceivedTotal,
				Received:      transactionNeeds[j].Received,
				Unit:          transactionNeeds[j].Unit,
			}
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
			CreatorID:             h.Transactions[i].CreatorID,
			CreationDate:          h.Transactions[i].CreationDate,
			EventType:             HelpEventType,
			TransactionStatus:     h.Transactions[i].TransactionStatus,
			ResponderStatus:       h.Transactions[i].ResponderStatus,
			ReportURL:             h.Transactions[i].ReportURL,
			Receiver:              h.User.ToShortInfo(),
			Responder:             h.Transactions[i].Creator.ToShortInfo(),
			EventID:               h.ID,
			Comment:               h.Transactions[i].Comment,
			CompetitionDate:       h.Transactions[i].CompetitionDate.Time.Format(time.RFC3339),
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
