package handlers

import (
	service "Kurajj/internal/services"
	"github.com/gorilla/mux"
	"net/http"
)

type Handler struct {
	services *service.Service
}

func NewHandler(s *service.Service) Handler {
	return Handler{services: s}
}

func (h *Handler) InitRoutes() http.Handler {
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").Subrouter()

	var (
		//helpRequestRouter = apiRouter.PathPrefix("/help-request").Subrouter()
		//publicEvent       = apiRouter.PathPrefix("/public-event").Subrouter()
		proposalEvent = apiRouter.PathPrefix("/proposal-event").Subrouter()
	)
	// get all, get by id, delete, update, accept and declare, answer, get reports
	proposalEvent.HandleFunc("/").Methods(http.MethodGet)
}
