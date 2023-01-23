package handlers

import (
	"Kurajj/internal/handlers/models"
	httpHelper "Kurajj/pkg/http"
	"context"
	"net/http"
	"time"
)

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {

}

type userResponse struct {
	userID int
	err    error
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	user, err := models.UnmarshalNewUserInput(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
	}
	userch := make(chan userResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		userID, err := h.services.Authentication.SignUp(ctx, user)
		userch <- userResponse{
			userID: int(userID),
			err:    err,
		}
	}()

	for {
		select {
		case <-ctx.Done():
			httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "signing up new user took too long")
		case resp := <-userch:

		}
	}
}
