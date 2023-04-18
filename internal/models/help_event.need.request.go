package models

import "github.com/go-playground/validator/v10"

type NeedRequestCreate struct {
	Title  string `json:"title" validate:"required"`
	Amount int    `json:"amount" validate:"required,gte=0,lte=50"`
	Unit   Unit   `json:"unit" validate:"oneof=kilogram liter item"`
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
		ID:     0,
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
)
