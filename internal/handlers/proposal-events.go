package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type proposalEventCreateResponse struct {
	id  uint
	err error
}

// CreateProposalEvent creates a new proposal event
// @Summary      Create a new proposal event
// @Tags         Proposal Event
// @Accept       json
// @Produce      json
// @Param request body models.ProposalEventRequestCreate true "query params"
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/create [post]
func (h *Handler) CreateProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}
	event, err := models.UnmarshalProposalEventCreate(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	eventch := make(chan proposalEventCreateResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, err := h.services.ProposalEvent.CreateEvent(ctx, models.ProposalEvent{
			AuthorID:              userID.(uint),
			Title:                 event.Title,
			Description:           event.Description,
			CreationDate:          time.Now(),
			MaxConcurrentRequests: uint(event.MaxConcurrentRequests),
			RemainingHelps:        event.MaxConcurrentRequests,
		})

		eventch <- proposalEventCreateResponse{
			id:  id,
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "creating proposal event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		eventResponse := models.CreationResponse{ID: int(resp.id)}
		err := httpHelper.SendHTTPResponse(w, eventResponse)
		if err != nil {
			return
		}
	}
}

// UpdateProposalEvent updates a proposal event
// @Summary      Update proposal event
// @Tags         Proposal Event
// @Accept       json
// @Produce      json
// @Param request body models.ProposalEventRequestUpdate true "query params"
// @Param        id   path int  true  "ID"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/update/{id} [put]
func (h *Handler) UpdateProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	event, err := models.UnmarshalProposalEventUpdate(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	id, ok := mux.Vars(r)["id"]
	parsedID, err := strconv.Atoi(id)
	if !ok || err != nil {
		response := "there is no id for updating proposal event in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		err = h.services.ProposalEvent.UpdateEvent(ctx, models.ProposalEvent{
			ID:          uint(parsedID),
			Title:       event.Title,
			Description: event.Description,
			CompetitionDate: sql.NullTime{
				Time: event.CompetitionDate,
			},
			Category: event.Category,
		})

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "creating proposal event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		if err != nil {
			return
		}
	}
}

// DeleteProposalEvent deletes a proposal event
// @Summary      Delete proposal event
// @Tags         Proposal Event
// @Accept       json
// @Produce      json
// @Param        id   path int  true  "ID"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/delete/{id} [delete]
func (h *Handler) DeleteProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, ok := mux.Vars(r)["id"]
		parsedID, err := strconv.Atoi(id)
		if !ok || err != nil {
			response := "there is no id for delete in URL"
			if err != nil {
				response = err.Error()
			}
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
			return
		}
		err = h.services.ProposalEvent.DeleteEvent(ctx, uint(parsedID))

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "deleting proposal event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type getProposalEvent struct {
	proposalEvent models.ProposalEventGetResponse
	err           error
}

// GetProposalEvent gets proposal event by id
// @Summary      Get proposal event by id
// @Tags         Proposal Event
// @Accept       json
// @Produce      json
// @Param        id   path int  true  "ID"
// @Success      200  {object} models.ProposalEventGetResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/get/{id} [get]
func (h *Handler) GetProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	eventch := make(chan getProposalEvent)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, ok := mux.Vars(r)["id"]
		parsedID, err := strconv.Atoi(id)
		if !ok || err != nil {
			response := "there is no id for getting in URL"
			if err != nil {
				response = err.Error()
			}
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
			return
		}
		event, err := h.services.ProposalEvent.GetEvent(ctx, uint(parsedID))

		eventch <- getProposalEvent{
			proposalEvent: models.GetProposalEvent(event),
			err:           err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting proposal event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		httpHelper.SendHTTPResponse(w, resp.proposalEvent)
	}
}

type getProposalEvents struct {
	proposalEvents models.ProposalEvents
	err            error
}

// GetProposalEvents gets all proposal events
// @Summary      Get all proposal events
// @Tags         Proposal Event
// @Accept       json
// @Produce      json
// @Success      200  {object} models.ProposalEvents
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/get [get]
func (h *Handler) GetProposalEvents(w http.ResponseWriter, r *http.Request) {
	// TODO add filters
	// TODO add sorts
	defer r.Body.Close()

	eventch := make(chan getProposalEvents)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		events, err := h.services.ProposalEvent.GetEvents(ctx)

		eventch <- getProposalEvents{
			proposalEvents: models.GetProposalEvents(events...),
			err:            err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting proposal event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		httpHelper.SendHTTPResponse(w, resp.proposalEvents)
	}
}

// GetUsersProposalEvents get all proposal events created by user requester id
// @Summary      Get all proposal events created by user requester id
// @Tags         Proposal Event
// @Accept       json
// @Produce      json
// @Success      200  {object} models.ProposalEvents
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/get-own [get]
func (h *Handler) GetUsersProposalEvents(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}
	eventch := make(chan getProposalEvents)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		events, err := h.services.ProposalEvent.GetUserProposalEvents(ctx, userID.(uint))

		eventch <- getProposalEvents{
			proposalEvents: models.GetProposalEvents(events...),
			err:            err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting user's proposal event took too long")
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		httpHelper.SendHTTPResponse(w, resp.proposalEvents)
	}
}

func (h *Handler) SendProposalEventComplaint(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetProposalEventReports(w http.ResponseWriter, r *http.Request) {

}

// TODO add comment CRUD
// TODO add report CRUD
// TODO add Transaction logic

// ResponseProposalEvent creates new transaction with waiting status for the proposal event if slot is available
// @Summary      Create new transaction with waiting status for the proposal event if slot is available
// @Tags         Proposal Event
// @Accept       json
// @Param        id   path int  true  "ID"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/response [post]
func (h *Handler) ResponseProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}
	errch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, ok := mux.Vars(r)["id"]
		parsedID, err := strconv.Atoi(id)
		if !ok || err != nil {
			response := "there is no id for getting in URL"
			if err != nil {
				response = err.Error()
			}
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
			return
		}
		err = h.validateProposalEventTransactionRequest(ctx, uint(parsedID))
		if err != nil {
			errch <- errResponse{
				err: err,
			}

			return
		}
		err = h.services.ProposalEvent.Response(ctx, uint(parsedID), userID.(uint))

		errch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "responding proposal event took too long")
		return
	case resp := <-errch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) validateProposalEvent() {

}

// AcceptProposalEventResponse updates proposal event transaction's status to models.InProcess state
// @Summary      Update proposal event transaction's status to models.InProcess state
// @Tags         Proposal Event
// @Accept       json
// @Param        id   path int  true  "ID"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/accept [post]
func (h *Handler) AcceptProposalEventResponse(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	errch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, ok := mux.Vars(r)["id"]
		parsedID, err := strconv.Atoi(id)
		if !ok || err != nil {
			response := "there is no id for getting in URL"
			if err != nil {
				response = err.Error()
			}
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
			return
		}
		err = h.services.ProposalEvent.Accept(ctx, uint(parsedID))

		errch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "responding proposal event took too long")
		return
	case resp := <-errch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) validateProposalEventTransactionRequest(ctx context.Context, transactionID uint) error {
	transaction, err := h.services.ProposalEvent.GetEvent(ctx, transactionID)
	if err != nil {
		return err
	}

	if transaction.RemainingHelps-1 < 0 {
		return fmt.Errorf("there is no available slot")
	}
	return nil
}

// UpdateProposalEventTransactionStatus updates proposal event transaction's status to one of models.Status state
// @Summary      Update proposal event transaction's status to to one of models.Status state
// @Tags         Proposal Event
// @Accept       json
// @Param        id   path int  true  "ID"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/update-status/{id} [post]
func (h *Handler) UpdateProposalEventTransactionStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	transactionID, ok := mux.Vars(r)["transactionID"]
	parsedTransactionID, err := strconv.Atoi(transactionID)
	if !ok || err != nil {
		response := "there is no transactionID for getting in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	userID := r.Context().Value("transactionID")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user transactionID isn't in context")
		return
	}
	s, err := models.UnmarshalStatusExport(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		err := h.services.ProposalEvent.UpdateStatus(ctx, s.Status, uint(parsedTransactionID), userID.(uint))

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, fmt.Sprintf("updating proposal event transaction wtih transactionID - %d took too long", parsedTransactionID))
		return
	case resp := <-eventch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
