package handlers

import (
	_ "Kurajj/docs"
	service "Kurajj/internal/services"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
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

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

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

	eventsSubRouter := apiRouter.PathPrefix("/events").Subrouter()
	proposalEventSubRouter := eventsSubRouter.PathPrefix("/proposal").Subrouter()

	proposalEventSubRouter.HandleFunc("/create", h.CreateProposalEvent).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/update/{id}", h.UpdateProposalEvent).
		Methods(http.MethodPut, http.MethodPatch)
	proposalEventSubRouter.HandleFunc("/get/{id}", h.GetProposalEvent).
		Methods(http.MethodGet)
	proposalEventSubRouter.HandleFunc("/get", h.GetProposalEvents).
		Methods(http.MethodGet)
	proposalEventSubRouter.HandleFunc("/get-own", h.GetUsersProposalEvents).
		Methods(http.MethodGet)
	proposalEventSubRouter.HandleFunc("/delete/{id}", h.DeleteProposalEvent).
		Methods(http.MethodDelete)
	proposalEventSubRouter.HandleFunc("/reports/{id}", h.GetProposalEventReports).
		Methods(http.MethodGet)
	proposalEventSubRouter.HandleFunc("/complain/{id}", h.SendProposalEventComplaint).
		Methods(http.MethodPost)

	proposalEventSubRouter.HandleFunc("/comments/{id}", h.GetCommentsInProposalEvent).
		Methods(http.MethodGet)
	proposalEventSubRouter.HandleFunc("/comment", h.WriteCommentInProposalEvent).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/comment/{id}", h.UpdateProposalEventComment).
		Methods(http.MethodPut)
	proposalEventSubRouter.HandleFunc("/comment/{id}", h.DeleteProposalEventComment).
		Methods(http.MethodDelete)
	proposalEventSubRouter.HandleFunc("/transactions/{id}", h.GetProposalEventTransactions).
		Methods(http.MethodGet)

	proposalEventSubRouter.HandleFunc("/response/{id}", h.ResponseProposalEvent).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/accept/{id}", h.AcceptProposalEventResponse).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/update-status/{id}", h.UpdateProposalEventTransactionStatus).
		Methods(http.MethodPost)

	return r
}
