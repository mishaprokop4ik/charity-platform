package models

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"io"
)

const defaultImagePath = "https://charity-platform.s3.amazonaws.com/images/volunteer-care-old-people-nurse-isolated-young-human-helping-senior-volunteers-service-helpful-person-nursing-elderly-decent-vector-set_53562-17770.avif"

type HelpEventCreateRequest struct {
	Title       string              `json:"title" validate:"required"`
	Description string              `json:"description" validate:"required"`
	Needs       []NeedRequestCreate `json:"needs" validate:"required"`
	FileBytes   []byte              `json:"fileBytes"`
	FileType    string              `json:"fileType"`
	Tags        []TagRequestCreate  `json:"tags"`
}

func (h *HelpEventCreateRequest) validateFile(fl validator.FieldLevel) bool {
	fileBytes := fl.Field().Interface().([]byte)
	fileType := fl.Parent().Elem().FieldByName("FileType").String()
	if len(fileBytes) == 0 && fileType == "" {
		return true
	}
	return len(fileBytes) > 0 && fileType != ""
}

func (h *HelpEventCreateRequest) Validate() error {
	for i, n := range h.Needs {
		if n.Unit == "" {
			h.Needs[i].Unit = Item
		}
		if err := n.Validate(); err != nil {
			return err
		}
	}

	helpEventValidator := validator.New()
	helpEventValidator.RegisterValidation("fileFields", h.validateFile)
	if err := helpEventValidator.Struct(h); err != nil {
		return err
	}

	return nil
}

func (h *HelpEventCreateRequest) ToInternal() *HelpEvent {
	needs := make([]Need, len(h.Needs))
	for i, n := range h.Needs {
		needs[i] = n.ToInternal()
	}
	event := &HelpEvent{
		Title:       h.Title,
		Description: h.Description,
		Needs:       needs,
	}
	if len(h.FileBytes) == 0 {
		event.ImagePath = defaultImagePath
	}

	return event
}

func NewHelpEventCreateRequest(reader *io.ReadCloser) (*HelpEventCreateRequest, error) {
	event := &HelpEventCreateRequest{}
	decoder := json.NewDecoder(*reader)
	err := decoder.Decode(&event)

	return event, err
}
