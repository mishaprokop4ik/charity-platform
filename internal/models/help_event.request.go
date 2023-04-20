package models

import (
	"bytes"
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

func validateFile(fl validator.FieldLevel) bool {
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
	helpEventValidator.RegisterValidation("fileFields", validateFile)
	if err := helpEventValidator.Struct(h); err != nil {
		return err
	}

	return nil
}

func (h *HelpEventCreateRequest) ToInternal(authorID uint) *HelpEvent {
	needs := make([]Need, len(h.Needs))
	for i, n := range h.Needs {
		needs[i] = n.ToInternal()
	}
	event := &HelpEvent{
		Title:       h.Title,
		Description: h.Description,
		Needs:       needs,
		Status:      InActive,
		CreatedBy:   authorID,
	}
	if len(h.FileBytes) == 0 || h.FileType == "" {
		event.ImagePath = defaultImagePath
	} else if len(h.FileBytes) != 0 && h.FileType != "" {
		event.FileType = h.FileType
		event.File = bytes.NewBuffer(h.FileBytes)
	}

	tags := make([]Tag, len(h.Tags))
	for i := 0; i < len(h.Tags); i++ {

		tagValues := make([]TagValue, len(h.Tags[i].Values))
		for j, value := range h.Tags[i].Values {
			tagValues[j] = TagValue{
				Value: value,
			}
		}

		tags[i] = Tag{
			Title:     h.Tags[i].Title,
			EventType: HelpEventType,
			Values:    tagValues,
		}
	}

	event.Tags = tags

	return event
}

func NewHelpEventCreateRequest(reader *io.ReadCloser) (*HelpEventCreateRequest, error) {
	event := &HelpEventCreateRequest{}
	decoder := json.NewDecoder(*reader)
	err := decoder.Decode(&event)

	return event, err
}
