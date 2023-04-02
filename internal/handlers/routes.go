package handlers

import (
	_ "Kurajj/docs"
	service "Kurajj/internal/services"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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

	openAPI := r.PathPrefix("/open-api").Subrouter()
	openAPI.Use(h.SetId)
	openAPI.HandleFunc("/proposal-search", h.SearchProposalEvents).
		Methods(http.MethodPost)
	openAPI.HandleFunc("/proposal/{id}", h.GetProposalEvent).
		Methods(http.MethodGet)
	openAPI.HandleFunc("/proposal/", h.GetProposalEvents).
		Methods(http.MethodGet)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(h.Authentication)

	apiRouter.HandleFunc("/refresh-user-data", h.RefreshUserData).Methods(http.MethodPost)
	apiRouter.HandleFunc("/read-notifications", h.ReadNotifications).Methods(http.MethodPut)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/sign-up", h.UserSignUp).
		Methods(http.MethodPost)
	auth.HandleFunc("/sign-in", h.UserSignIn).
		Methods(http.MethodPost)
	auth.HandleFunc("/confirm/{email}", h.ConfirmEmail)
	auth.HandleFunc("/sign-in-admin", h.AdminSignIn).
		Methods(http.MethodPost)
	auth.HandleFunc("/refresh-token", h.RefreshTokens).Methods(http.MethodPost)

	adminSubRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminSubRouter.HandleFunc("/create", h.CreateNewAdmin)

	eventsSubRouter := apiRouter.PathPrefix("/events").Subrouter()
	proposalEventSubRouter := eventsSubRouter.PathPrefix("/proposal").Subrouter()

	proposalEventSubRouter.HandleFunc("/create", h.CreateProposalEvent).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/update/{id}", h.UpdateProposalEvent).
		Methods(http.MethodPut, http.MethodPatch)
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

	proposalEventSubRouter.HandleFunc("/response", h.ResponseProposalEvent).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/accept/{id}", h.AcceptProposalEventResponse).
		Methods(http.MethodPost)
	proposalEventSubRouter.HandleFunc("/update-status/{id}", h.UpdateProposalEventTransactionStatus).
		Methods(http.MethodPost)

	proposalEventSubRouter.HandleFunc("/tags/{id}", h.GetProposalEventTags).Methods(http.MethodGet)

	tags := apiRouter.PathPrefix("/tags").Subrouter()
	tags.HandleFunc("/upsert", h.UpsertTags).Methods(http.MethodPost)
	tags.HandleFunc("/user-search", h.UpsertUserSearch).Methods(http.MethodPost)

	handler := cors.AllowAll().Handler(r)

	return handler
}
