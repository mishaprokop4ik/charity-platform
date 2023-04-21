package models

import (
	"bytes"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"io"
	"time"
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
		Status:      Active,
		CreatedBy:   authorID,
	}
	if len(h.FileBytes) == 0 || h.FileType == "" {
		event.ImagePath = defaultImagePath
	} else if len(h.FileBytes) != 0 && h.FileType != "" {
		event.FileType = h.FileType
		event.File = bytes.NewBuffer(h.FileBytes)
	}

	location := Address{}
	for i, t := range h.Tags {
		if t.Title == "location" && len(t.Values) >= DecodedAddressLength {
			if len(t.Values[0]) != 0 {
				location.Region = t.Values[0]
			}
			if len(t.Values[1]) != 0 {
				location.City = t.Values[1]
			}
			if len(t.Values[2]) != 0 {
				location.District = t.Values[2]
			}
			if len(t.Values[3]) != 0 {
				location.HomeLocation = t.Values[3]
			}
			location.EventType = HelpEventType
			h.Tags = append(h.Tags[:i], h.Tags[i+1:]...)
		}
	}

	event.Tags = h.TagsInternal()

	return event
}

func (h *HelpEventCreateRequest) TagsInternal() []Tag {
	tags := make([]Tag, len(h.Tags))
	for i, tag := range h.Tags {
		tagValues := make([]TagValue, len(tag.Values))
		for _, tagValue := range tag.Values {
			tagValues[i] = TagValue{
				Value: tagValue,
			}
		}
		tags[i] = Tag{
			Title:     tag.Title,
			EventType: HelpEventType,
			Values:    tagValues,
		}
	}
	return tags
}

func NewHelpEventCreateRequest(reader *io.ReadCloser) (*HelpEventCreateRequest, error) {
	event := &HelpEventCreateRequest{}
	decoder := json.NewDecoder(*reader)
	err := decoder.Decode(&event)

	return event, err
}

type HelpEventsResponse struct {
	Events []HelpEventResponse `json:"events"`
}

type HelpEventsItems struct {
	HelpEvents []HelpEventResponse `json:"items"`
}

type HelpEventsWithPagination struct {
	HelpEventsItems
	Pagination
}

func CreateHelpEventsResponse(events []HelpEvent) HelpEventsResponse {
	response := HelpEventsResponse{
		Events: make([]HelpEventResponse, len(events)),
	}

	for i := range events {
		response.Events[i] = events[i].Response()
	}

	return response
}

func (h *HelpEventsResponse) Bytes() []byte {
	bytes, _ := json.Marshal(h)
	return bytes
}

type HelpEventRequestUpdate struct {
	ID              uint        `json:"id"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	CompetitionDate time.Time   `json:"competitionDate"`
	Status          EventStatus `json:"status"`
	FileBytes       []byte      `json:"fileBytes"`
	FileType        string      `json:"fileType"`
}

func UnmarshalHelpEventUpdate(r *io.ReadCloser) (HelpEventRequestUpdate, error) {
	e := HelpEventRequestUpdate{}
	err := json.NewDecoder(*r).Decode(&e)
	return e, err
}

func (p *HelpEventRequestUpdate) Internal() HelpEvent {
	event := HelpEvent{
		ID:             p.ID,
		Title:          p.Title,
		Description:    p.Description,
		Status:         p.Status,
		CompletionTime: p.CompetitionDate,
	}
	if len(p.FileBytes) != 0 {
		event.File = bytes.NewReader(p.FileBytes)
		event.FileType = p.FileType
	}

	return event
}
