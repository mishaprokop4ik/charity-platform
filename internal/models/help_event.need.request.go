package models

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"time"
)

type NeedRequestCreate struct {
	Title  string  `json:"title" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gte=0,lte=50"`
	Unit   Unit    `json:"unit" validate:"oneof=kilogram liter item work"`
}

func (n *NeedRequestCreate) Validate() error {
	if n.Unit == "" {
		n.Unit = Item
	}

	validate := validator.New()

	return validate.Struct(n)
}

func (n *NeedRequestCreate) ToInternal() Need {
	return Need{
		Title:  n.Title,
		Amount: n.Amount,
		Unit:   n.Unit,
	}
}

type Unit string

const (
	Kilogram = "kilogram"
	Liter    = "liter"
	Item     = "item"
	Work     = "work"
)

type NeedResponse struct {
	ID            uint    `json:"id"`
	Title         string  `json:"title"`
	Amount        float64 `json:"amount"`
	ReceivedTotal float64 `json:"receivedTotal"`
	Received      float64 `json:"received"`
	Unit          Unit    `json:"unit"`
}

type NeedTransactionUpdateRequest struct {
	ID       uint    `json:"id"`
	Title    string  `json:"title"`
	Received float64 `json:"received"`
	Unit     Unit    `json:"unit"`
}

type HelpEventResponse struct {
	ID                    uint                           `json:"id"`
	Title                 string                         `json:"title"`
	Description           string                         `json:"description"`
	CreationDate          time.Time                      `json:"creationDate"`
	CompetitionDate       time.Time                      `json:"competitionDate"`
	Status                EventStatus                    `json:"status"`
	EndDate               time.Time                      `json:"endDate"`
	ImageURL              string                         `json:"imageURL"`
	AuthorInfo            UserShortInfo                  `json:"authorInfo"`
	Comments              []CommentResponse              `json:"comments"`
	Transactions          []HelpEventTransactionResponse `json:"transactions"`
	Tags                  []TagResponse                  `json:"tags"`
	Needs                 []NeedResponse                 `json:"needs"`
	CompletionPercentages float64                        `json:"completionPercentages"`
}

func (p HelpEventResponse) Bytes() []byte {
	bytes, _ := json.Marshal(p)
	return bytes
}
