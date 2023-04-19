package models

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"io"
	"time"
)

type HelpEventTransactionUpdateRequest struct {
	ID        uint                           `json:"id"`
	FileBytes []byte                         `json:"fileBytes"`
	FileType  string                         `json:"fileType"`
	Status    TransactionStatus              `json:"status" validate:"required"`
	Needs     []NeedTransactionUpdateRequest `json:"needs"`
}

func (h *HelpEventTransactionUpdateRequest) needsToInternal(isEventCreator bool, eventID uint) []Need {
	needs := make([]Need, len(h.Needs))
	for i := range h.Needs {
		need := Need{
			ID:            h.Needs[i].ID,
			Title:         h.Needs[i].Title,
			TransactionID: &h.ID,
			HelpEventID:   eventID,
			Unit:          h.Needs[i].Unit,
		}
		if isEventCreator {
			need.ReceivedTotal = h.Needs[i].Received
		} else {
			need.Received = h.Needs[i].Received
		}
		needs[i] = need
	}

	return needs
}

func (h *HelpEventTransactionUpdateRequest) ToInternal(eventCreator bool, helpEventID ID, requesterID uint) HelpEventTransaction {
	transaction := HelpEventTransaction{
		Needs:           h.needsToInternal(eventCreator, uint(helpEventID)),
		CompetitionDate: time.Now(),
		HelpEventID:     (*uint)(&helpEventID),
		TransactionID:   &h.ID,
	}
	transaction.EventCreator = eventCreator
	if eventCreator {
		transaction.CompetitionDate = time.Now()
		transaction.TransactionStatus = h.Status
		transaction.HelpEventCreatorID = requesterID
	} else {
		transaction.ResponderStatus = h.Status
		transaction.TransactionCreatorID = requesterID
	}
	return transaction
}

func (h *HelpEventTransactionUpdateRequest) Validate() error {
	helpEventValidator := validator.New()
	err := helpEventValidator.RegisterValidation("fileFields", validateFile)
	if err != nil {
		return err
	}
	if err := helpEventValidator.Struct(h); err != nil {
		return err
	}
	if h.Status == Completed && len(h.FileBytes) == 0 || h.FileType == "" {
		return fmt.Errorf("requires file when status is changed to completed")
	}

	return nil
}

func NewHelpEventTransactionUpdateRequest(r *io.ReadCloser) (HelpEventTransactionUpdateRequest, error) {
	helpEventTransactionUpdateRequest := HelpEventTransactionUpdateRequest{}
	err := json.NewDecoder(*r).Decode(&helpEventTransactionUpdateRequest)
	if err != nil {
		return HelpEventTransactionUpdateRequest{}, err
	}

	lo.Filter(helpEventTransactionUpdateRequest.Needs, func(need NeedTransactionUpdateRequest, index int) bool {
		return need.Received == 0
	})

	err = helpEventTransactionUpdateRequest.Validate()

	return helpEventTransactionUpdateRequest, err
}
