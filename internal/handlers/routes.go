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
	apiRouter.Use(h.Authentication)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/sign-up", h.UserSignUp).
		Methods(http.MethodPost)
	auth.HandleFunc("/sign-in", h.UserSignIn).
		Methods(http.MethodPost)
	auth.HandleFunc("/confirm/{email}", h.ConfirmEmail)
	auth.HandleFunc("/sign-in-admin", h.AdminSignIn).
		Methods(http.MethodPost)

	adminSubRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminSubRouter.HandleFunc("/create", h.CreateNewAdmin)

	return r
}
