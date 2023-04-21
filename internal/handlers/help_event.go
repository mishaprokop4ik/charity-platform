package handlers

import (
	"Kurajj/internal/models"
	"Kurajj/internal/models/search"
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) initHelpEventHandlers(events *mux.Router) {
	helpEvent := events.PathPrefix("/help").Subrouter()
	helpEvent.HandleFunc("/create", h.handleCreateHelpEvent).Methods(http.MethodPost)
	helpEvent.HandleFunc("/response", h.handleApplyTransaction).Methods(http.MethodPost)
	helpEvent.HandleFunc("/transaction", h.handleUpdateTransactionResponseHelpEvent).Methods(http.MethodPut)
	helpEvent.HandleFunc("/own", h.handleGetOwnHelpEvents).Methods(http.MethodGet)
	helpEvent.HandleFunc("/{id}", h.handleUpdateHelpEvent).Methods(http.MethodPut)
	helpEvent.HandleFunc("/comment", h.handleWriteCommentInHelpEvent).Methods(http.MethodPost)
	helpEvent.HandleFunc("/comment/{id}", h.handleUpdateHelpEventComment).Methods(http.MethodPut)
	helpEvent.HandleFunc("/comment/{id}", h.handleDeleteHelpEventComment).Methods(http.MethodDelete)
	helpEvent.HandleFunc("/comments/{id}", h.handleGetCommentsInHelpEvent).Methods(http.MethodGet)
	helpEvent.HandleFunc("/statistics", h.handleGetHelpEventStatistics).Methods(http.MethodGet)
}

// GetHelpEventByID gets help event by id
// @Summary      Get help event by id
// @Tags         Help Event
// @SearchValuesResponse         Help Event
// @Accept       json
// @Produce      json
// @Param        id   path int  true  "ID"
// @Success      200  {object} models.HelpEventResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /open-api/help/{id} [get]
func (h *Handler) handleGetHelpEventByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	eventch := make(chan getHelpEvent)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
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
	go func() {
		event, err := h.services.HelpEvent.GetHelpEventByID(ctx, models.ID(parsedID))

		eventch <- getHelpEvent{
			helpEvent: event,
			err:       err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting help event took too long")
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
		err := httpHelper.SendHTTPResponse(w, resp.helpEvent.Response())
		if err != nil {
			zlog.Log.Error(err, "got an error")
		}
	}
}

func (h *Handler) GetUserHelpEvents(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) SearchHelpEvents(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetHelpEvents(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// CreateHelpEvent creates a new help event
// @Summary      Create a new Help event
// @Tags         Help Event
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
func (h *Handler) handleCreateHelpEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	event, err := models.NewHelpEventCreateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = event.Validate(); err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := r.Context().Value(MemberIDContextKey)
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}

	eventch := make(chan idResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, err := h.services.HelpEvent.CreateHelpEvent(ctx, event.ToInternal(userID.(uint)))

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

func (h *Handler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

// handleUpdateHelpEvent updates a help event
// @Summary      Update help event
// @SearchValuesResponse         Help Event
// @Tags         Help Event
// @Accept       json
// @Produce      json
// @Param request body models.HelpEventRequestUpdate true "query params"
// @Param        id   path int  true  "ID"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/update/{id} [put]
func (h *Handler) handleUpdateHelpEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	event, err := models.UnmarshalHelpEventUpdate(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	id, ok := mux.Vars(r)["id"]
	parsedID, err := strconv.Atoi(id)
	if !ok || err != nil {
		response := "there is no id for updating help event in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	event.ID = uint(parsedID)

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		err = h.services.HelpEvent.UpdateEvent(ctx, event.Internal())

		eventch <- errResponse{
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
		w.WriteHeader(http.StatusOK)
		if err != nil {
			return
		}
	}
}

// handleSearchHelpEvents gets models.HelpEventsResponse by given order and filter values
// @Param        id   path int  true  "ID"
// @Summary      Return help events by given order and filter values
// @Param request body search.AllEventsSearch true "query params"
// @SearchValuesResponse         Help Event
// @Tags         Help Event
// @Accept       json
// @Success      200  {object}  models.HelpEventsWithPagination
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /open-api/help-search [post]
func (h *Handler) handleSearchHelpEvents(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	searchValues, err := search.UnmarshalAllEventsSearch(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := r.Context().Value(MemberIDContextKey)
	searchValuesInternal := searchValues.Internal()
	if userID != nil {
		userIDParsed, ok := userID.(uint)
		if !ok {
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
			return
		}
		searchValuesInternal.SearcherID = &userIDParsed
	}
	eventch := make(chan getHelpEventPagination)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		events, respError := h.services.HelpEvent.GetHelpEventBySearch(ctx, models.HelpSearchInternal(searchValuesInternal))

		eventch <- getHelpEventPagination{
			resp: models.HelpEventsWithPagination{
				HelpEventsItems: models.GetHelpEventItems(events.Events...),
				Pagination:      events.Pagination,
			},
			err: respError,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting all transactions for help event took too long")
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

		w.WriteHeader(http.StatusOK)
		err = httpHelper.SendHTTPResponse(w, resp.resp)
		if err != nil {
			zlog.Log.Error(err, "could not send help events")
		}
	}
}

// handleUpdateTransactionResponseHelpEvent updates transaction status and if requester is a creator of event updates event.
// @Summary      Update transaction status and if requester is a creator of event updates event.
// @Tags         Help Event
// @Accept       json
// @Produce      json
// @Param request body models.HelpEventTransactionUpdateRequest true "query params"
// @Tags         Help Event
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/transaction [put]
func (h *Handler) handleUpdateTransactionResponseHelpEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	transaction, err := models.NewHelpEventTransactionUpdateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	oldTransaction, err := h.services.Transaction.GetTransactionByID(ctx, transaction.ID)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if oldTransaction.TransactionStatus == models.Completed {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest,
			"transaction's status cannot be changed when it is already completed")
		return
	}

	helpEvent, err := h.services.HelpEvent.GetHelpEventByTransactionID(ctx, models.ID(transaction.ID))
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot get help event by requested transaction %d id",
			transaction.ID))
		return
	}

	userID := r.Context().Value(MemberIDContextKey)
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}

	eventCreator := userID.(uint) == helpEvent.CreatedBy
	eventch := make(chan errResponse)
	go func() {
		err := h.services.HelpEvent.UpdateTransactionStatus(ctx, transaction.ToInternal(eventCreator, models.ID(helpEvent.ID), userID.(uint)),
			bytes.NewReader(transaction.FileBytes), transaction.FileType)

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "applying took too long")
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
		w.WriteHeader(http.StatusOK)
	}
}

// handleApplyTransaction creates a new help event transaction with TransactionStatus - models.Waiting, ResponderStatus - models.NotStarted.
// @Description  Create a new help event transaction with TransactionStatus - waiting, ResponderStatus - not_started.
// @Summary      Create a new help event transaction with TransactionStatus - waiting, ResponderStatus - not_started.
// @Tags         Help Event
// @Accept       json
// @Produce      json
// @Param request body models.TransactionAcceptCreateRequest true "query params"
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/response [post]
func (h *Handler) handleApplyTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}
	eventch := make(chan idResponse)
	transactionInfo, err := models.UnmarshalTransactionAcceptCreateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		transactionID, err := h.services.HelpEvent.CreateRequest(ctx, models.ID(userID.(uint)), transactionInfo)

		eventch <- idResponse{
			id:  int(transactionID),
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "applying took too long")
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
		httpHelper.SendHTTPResponse(w, models.CreationResponse{ID: resp.id})
	}
}

// handleGetOwnHelpEvents returns all help events created by user.
// @Summary      Return all help events created by user.
// @Tags         Help Event
// @Accept       json
// @Produce      json
// @Success      201  {object}  models.HelpEventsResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/own [get]
func (h *Handler) handleGetOwnHelpEvents(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	eventch := make(chan helpEventsResponse)
	defer cancel()
	go func() {
		events, err := h.services.HelpEvent.GetUserHelpEvents(ctx, models.ID(userID.(uint)))

		eventch <- helpEventsResponse{
			events: events,
			err:    err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting help events took too long")
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

		helpEventsResponse := models.CreateHelpEventsResponse(resp.events)

		httpHelper.SendHTTPResponse(w, &helpEventsResponse)
	}
}

// handleWriteCommentInHelpEvent creates new comment in help event
// @Param request body models.CommentCreateRequest true "query params"
// @Tags         Help Event
// @Summary      Create new comment in help event
// @SearchValuesResponse         Help Event
// @Accept       json
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/comment [post]
func (h *Handler) handleWriteCommentInHelpEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user transactionID isn't in context")
		return
	}
	comment, err := models.UnmarshalCommentCreateRequest(&r.Body)
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
			EventType:    models.HelpEventType,
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
			case models.ErrNotFound.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		commentResponse := models.CreationResponse{ID: resp.id}
		_ = httpHelper.SendHTTPResponse(w, commentResponse)
	}
}

// handleGetCommentsInHelpEvent takes all comments in help event by its id
// @Summary      Take all comments in help event by its id
// @SearchValuesResponse         Help Event
// @Tags         Help Event
// @Accept       json
// @Param        id   path int  true  "ID"
// @Success      200  {object}  models.Comments
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/comments/id [get]
func (h *Handler) handleGetCommentsInHelpEvent(w http.ResponseWriter, r *http.Request) {
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
		comments, err := h.services.Comment.GetAllCommentsInEvent(ctx, uint(parsedEventID), models.HelpEventType)

		eventch <- commentsResponse{
			comments: comments,
			err:      err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout,
			fmt.Sprintf("updating help event transaction with eventID - %d took too long",
				parsedEventID))
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
		responseComments := models.Comments{
			Comments: make([]models.CommentResponse, len(resp.comments)),
		}

		for i, c := range resp.comments {
			user, err := h.services.Authentication.GetUserShortInfo(ctx, c.UserID)
			if err != nil {
				httpHelper.SendErrorResponse(w,
					http.StatusRequestTimeout,
					fmt.Sprintf("can not get user info, help event %s",
						err))
				return
			}
			updatedAt := ""
			if c.UpdatedAt.Valid {
				updatedAt = c.UpdatedAt.Time.String()
			}
			responseComments.Comments[i] = models.CommentResponse{
				ID:            c.ID,
				Text:          c.Text,
				CreationDate:  c.CreationDate,
				IsUpdated:     c.IsUpdated,
				UpdateTime:    updatedAt,
				UserShortInfo: user,
			}
		}

		_ = httpHelper.SendHTTPResponse(w, responseComments)
	}
}

// handleUpdateHelpEventComment updates help event comment
// @Param        id   path int  true  "ID"
// @Param request body models.CommentUpdateRequest true "query params"
// @Summary      Update help event comment
// @SearchValuesResponse         Help Event
// @Tags         Help Event
// @Accept       json
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/comment/{id} [put]
func (h *Handler) handleUpdateHelpEventComment(w http.ResponseWriter, r *http.Request) {
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

	comment, err := models.UnmarshalCommentUpdateRequest(&r.Body)
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
			EventType: models.HelpEventType,
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
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout,
			fmt.Sprintf("updating help event transaction with commentID - %d took too long",
				parsedCommentID))
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
		w.WriteHeader(http.StatusOK)
	}
}

// handleDeleteHelpEventComment deletes help event comment
// @Param        id   path int  true  "ID"
// @Summary      Update Help event comment
// @SearchValuesResponse         Help Event
// @Tags         Help Event
// @Accept       json
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/comment/{id} [delete]
func (h *Handler) handleDeleteHelpEventComment(w http.ResponseWriter, r *http.Request) {
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
		httpHelper.SendErrorResponse(w,
			http.StatusRequestTimeout,
			fmt.Sprintf("updating help event transaction with commentID - %d took too long",
				parsedCommentID))
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
		w.WriteHeader(http.StatusOK)
	}
}

// handleGetHelpEventStatistics takes statistics of help event from time.Now() - 28
// @Summary      Take statistics of help event from current date - 28 to current date
// @SearchValuesResponse         Help Event
// @Tags         Help Event
// @Success      200  {object}  models.HelpEventStatistics
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/statistics [get]
func (h *Handler) handleGetHelpEventStatistics(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var userIDParsed uint
	userID := r.Context().Value("id")
	if userID != nil {
		_, ok := userID.(uint)
		if !ok {
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
			return
		}
		userIDParsed = userID.(uint)
	}

	eventch := make(chan geHelpStatistics)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		statistics, err := h.services.HelpEvent.GetStatistics(ctx, 28, userIDParsed)

		eventch <- geHelpStatistics{
			statistics: statistics,
			err:        err,
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
			case models.ErrNotFound.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		err := httpHelper.SendHTTPResponse(w, resp.statistics)
		if err != nil {
			zlog.Log.Error(err, "got an error")
		}
	}
}
