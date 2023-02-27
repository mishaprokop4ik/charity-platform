package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
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
	transactionID, ok := mux.Vars(r)["id"]
	parsedTransactionID, err := strconv.Atoi(transactionID)
	if !ok || err != nil {
		response := "there is no transactionID for getting in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	userID := r.Context().Value("id")
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

// WriteCommentInProposalEvent creates new comment in proposal event
// @Param request body models.CommentCreateRequest true "query params"
// @Summary      Create new comment in proposal event
// @Tags         Proposal Event
// @Accept       json
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/comment [post]
func (h *Handler) WriteCommentInProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user transactionID isn't in context")
		return
	}
	comment, err := models.UnmarshalCommentCreateRequest(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	eventch := make(chan idResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, err := h.services.Comment.WriteComment(ctx, models.Comment{
			EventID:      comment.EventID,
			EventType:    models.ProposalEventType,
			Text:         comment.Text,
			CreationDate: time.Now(),
			UserID:       userID.(uint),
		})

		eventch <- idResponse{
			id:  int(id),
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "writing comment took too long")
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
		commentResponse := models.CreationResponse{ID: resp.id}
		_ = httpHelper.SendHTTPResponse(w, commentResponse)
	}
}

type commentsResponse struct {
	comments []models.Comment
	err      error
}

// GetCommentsInProposalEvent takes all comments in proposal event by its id
// @Summary      Take all comments in proposal event by its id
// @Tags         Proposal Event
// @Accept       json
// @Param        id   path int  true  "ID"
// @Success      200  {object}  models.Comments
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/comments/id [get]
func (h *Handler) GetCommentsInProposalEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	eventID, ok := mux.Vars(r)["id"]
	parsedEventID, err := strconv.Atoi(eventID)
	if !ok || err != nil {
		response := "there is no eventID for getting in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	eventch := make(chan commentsResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		comments, err := h.services.Comment.GetAllCommentsInEvent(ctx, uint(parsedEventID), models.ProposalEventType)

		eventch <- commentsResponse{
			comments: comments,
			err:      err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, fmt.Sprintf("updating proposal event transaction wtih eventID - %d took too long", parsedEventID))
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
		responseComments := models.Comments{
			Comments: make([]models.CommentResponse, len(resp.comments)),
		}

		for i, c := range resp.comments {
			user, err := h.services.Authentication.GetUserShortInfo(ctx, c.UserID)
			if err != nil {
				httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, fmt.Sprintf("can not get user info, proposal event %s", err))
				return
			}
			responseComments.Comments[i] = models.CommentResponse{
				ID:           c.ID,
				Text:         c.Text,
				CreationDate: c.CreationDate,
				IsUpdated:    c.IsUpdated,
				UpdateTime:   c.UpdatedAt,
				UserComment:  user,
			}
		}

		_ = httpHelper.SendHTTPResponse(w, responseComments)
	}
}

// UpdateProposalEventComment updates proposal event comment
// @Param        id   path int  true  "ID"
// @Param request body models.CommentUpdateRequest true "query params"
// @Summary      Update proposal event comment
// @Tags         Proposal Event
// @Accept       json
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/comment/{id} [put]
func (h *Handler) UpdateProposalEventComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	commentID, ok := mux.Vars(r)["id"]
	parsedCommentID, err := strconv.Atoi(commentID)
	if !ok || err != nil {
		response := "there is no commentID for getting in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	comment, err := models.UnmarshalCommentUpdateRequest(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		err := h.services.Comment.UpdateComment(ctx, models.Comment{
			ID:        uint(parsedCommentID),
			EventType: models.ProposalEventType,
			Text:      comment.Text,
			IsUpdated: true,
			UpdatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
		})

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, fmt.Sprintf("updating proposal event transaction wtih commentID - %d took too long", parsedCommentID))
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

// DeleteProposalEventComment deletes proposal event comment
// @Param        id   path int  true  "ID"
// @Summary      Update proposal event comment
// @Tags         Proposal Event
// @Accept       json
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/comment/{id} [delete]
func (h *Handler) DeleteProposalEventComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	commentID, ok := mux.Vars(r)["id"]
	parsedCommentID, err := strconv.Atoi(commentID)
	if !ok || err != nil {
		response := "there is no commentID for getting in URL"
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
		err := h.services.Comment.DeleteComment(ctx, uint(parsedCommentID))

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, fmt.Sprintf("updating proposal event transaction wtih commentID - %d took too long", parsedCommentID))
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

type transactionsResponse struct {
	transactions []models.Transaction
	err          error
}

// GetProposalEventTransactions gets all proposal event transactions
// @Param        id   path int  true  "ID"
// @Summary      Get all proposal event transactions(finished, in process, etc)
// @Tags         Proposal Event
// @Accept       json
// @Success      200  {object}  models.TransactionsExport
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/proposal/transactions/{id} [get]
func (h *Handler) GetProposalEventTransactions(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	eventID, ok := mux.Vars(r)["id"]
	parsedEventID, err := strconv.Atoi(eventID)
	if !ok || err != nil {
		response := "there is no commentID for getting in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	eventch := make(chan transactionsResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		transactions, err := h.services.Transaction.GetAllEventTransactions(ctx, uint(parsedEventID), models.ProposalEventType)

		eventch <- transactionsResponse{
			transactions: transactions,
			err:          err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting all transactions for proposal event took too long")
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

		transactions := models.TransactionsExport{
			Transactions: make([]models.TransactionResponse, len(resp.transactions)),
		}

		for i, t := range resp.transactions {
			transaction := models.TransactionResponse{
				ID:                t.ID,
				CreatorID:         t.CreatorID,
				EventID:           t.EventID,
				Comment:           t.Comment,
				EventType:         t.EventType,
				TransactionStatus: t.TransactionStatus,
				ResponderStatus:   t.ResponderStatus,
			}
			if t.CompetitionDate.Valid {
				transaction.CompetitionDate = t.CompetitionDate.Time
			}
			transactions.Transactions[i] = transaction
		}
		w.WriteHeader(http.StatusOK)
		err = httpHelper.SendHTTPResponse(w, transactions)
		if err != nil {
			zlog.Log.Error(err, "could not send proposal event transaction")
		}
	}
}
