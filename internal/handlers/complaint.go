package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) initComplaintHandlers(api *mux.Router) {
	complaint := api.PathPrefix("/complaint").Subrouter()
	complaint.HandleFunc("/", h.handleCreateComplaint).Methods(http.MethodPost)
	complaint.HandleFunc("/", h.handleGetComplaints).Methods(http.MethodGet)
	complaint.HandleFunc("/{id}/eventbanned", h.handleBanEvent).Methods(http.MethodPost)
	complaint.HandleFunc("/{id}/userbanned", h.handleBanUser).Methods(http.MethodPost)
}

func (h *Handler) handleCreateComplaint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	complaint, err := models.NewComplaintCreateRequest(&r.Body)
	if err != nil {
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
		id, err := h.services.Complain(ctx, complaint.Internal(userID.(int)))

		eventch <- idResponse{
			id:  id,
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

func (h *Handler) handleGetComplaints(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	eventch := make(chan complaintsResponse)
	defer cancel()
	go func() {
		complaints, err := h.services.GetAll(ctx)

		eventch <- complaintsResponse{
			complaints: complaints,
			err:        err,
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

		helpEventsResponse := models.CreateComplaintsResponse(resp.complaints)

		httpHelper.SendHTTPResponse(w, &helpEventsResponse)
	}
}

func (h *Handler) handleBanEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	event, err := models.NewEventBanCreateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := r.Context().Value(MemberIDContextKey)
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user id isn't in context")
		return
	}

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		err := h.services.BanEvent(ctx, event.ID, event.Type)

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
	}
}

func (h *Handler) handleBanUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, ok := mux.Vars(r)["id"]
	parsedID, err := strconv.Atoi(id)
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
		err := h.services.BanUser(ctx, models.ID(parsedID))

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout,
			fmt.Sprintf("banning used took too long"))
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
