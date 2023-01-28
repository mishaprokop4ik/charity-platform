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
	//apiRouter := r.PathPrefix("/api").Subrouter()

	auth := r.PathPrefix("/auth").Subrouter()
	userAuth := auth.PathPrefix("/user").Subrouter()
	userAuth.HandleFunc("/sign-up", h.UserSignUp).
		Methods(http.MethodPost)
	userAuth.HandleFunc("/sign-in", h.UserSignIn).
		Methods(http.MethodPost)
	adminAuth := auth.PathPrefix("/admin").Subrouter()
	adminAuth.Use(h.Authentication)
	adminAuth.HandleFunc("/sign-up", h.UserSignUp).
		Methods(http.MethodPost)
	adminAuth.HandleFunc("/sign-in", h.UserSignIn).
		Methods(http.MethodPost)
	var (
	//helpRequestRouter = apiRouter.PathPrefix("/help-request").Subrouter()
	//publicEvent       = apiRouter.PathPrefix("/public-event").Subrouter()
	//proposalEvent = apiRouter.PathPrefix("/proposal-event").Subrouter()
	)

	return r
}
