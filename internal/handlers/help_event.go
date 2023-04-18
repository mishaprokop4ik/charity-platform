package handlers

import (
	"Kurajj/internal/models"
	service "Kurajj/internal/services"
	httpHelper "Kurajj/pkg/http"
	"context"
	"net/http"
	"time"
)

type helpEventHandlers struct {
	services *service.Service
}

func (h *helpEventHandlers) GetHelpEventByID(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) GetUserHelpEvents(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) SearchHelpEvents(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) GetHelpEvents(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// CreateHelpEvent creates a new help event
// @Summary      Create a new Help event
// @SearchValuesResponse         Proposal Event
// @Accept       json
// @Produce      json
// @Param request body models.HelpEventCreateRequest true "query params"
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/create [post]
func (h *helpEventHandlers) CreateHelpEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}
	event, err := models.NewHelpEventCreateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = event.Validate(); err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	eventch := make(chan idResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, err := h.services.HelpEvent.CreateHelpEvent(ctx, event.ToInternal())

		eventch <- idResponse{
			id:  int(id),
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "creating help event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.ErrNotFound.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		eventResponse := models.CreationResponse{ID: resp.id}
		err := httpHelper.SendHTTPResponse(w, eventResponse)
		if err != nil {
			return
		}
	}
}

func (h *helpEventHandlers) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) ResponseTransaction(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) AcceptTransaction(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *helpEventHandlers) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func newHelpEventHandlers(s *service.Service) *helpEventHandlers {
	return &helpEventHandlers{services: s}
}
