package models

import "encoding/json"

type EventComplaintsResponse struct {
	Complaints []ComplaintsResponse `json:"complaints"`
}

func (e EventComplaintsResponse) Bytes() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}

func CreateComplaintsResponse(complaints []ComplaintsResponse) EventComplaintsResponse {
	response := EventComplaintsResponse{
		Complaints: make([]ComplaintsResponse, len(complaints)),
	}

	for i := range complaints {
		response.Complaints[i] = complaints[i]
	}

	return response
}
