package handlers

import (
	service "Kurajj/internal/services"
	"github.com/gorilla/mux"
	"net/http"
)

type Handler struct {
	services *service.Service
}

func New(s *service.Service) Handler {
	return Handler{services: s}
}

func (h *Handler) InitRoutes() http.Handler {
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").Subrouter()

	auth := apiRouter.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/sign-up", h.SignUp).
		Methods(http.MethodPost)
	auth.HandleFunc("/sign-in", h.SignIn).
		Methods(http.MethodPost)

	var (
	//helpRequestRouter = apiRouter.PathPrefix("/help-request").Subrouter()
	//publicEvent       = apiRouter.PathPrefix("/public-event").Subrouter()
	//proposalEvent = apiRouter.PathPrefix("/proposal-event").Subrouter()
	)
}
