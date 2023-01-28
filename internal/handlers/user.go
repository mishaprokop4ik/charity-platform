package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
	"context"
	"net/http"
	"time"
)

type userSignInResponse struct {
	err  error
	resp models.SignedInUser
}

func (h *Handler) UserSignIn(w http.ResponseWriter, r *http.Request) {
	user, err := models.UnmarshalSignInUser(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	userch := make(chan userSignInResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		user, err := h.services.Authentication.SignIn(ctx, models.User{Email: string(user.Email), Password: user.Password})
		userch <- userSignInResponse{
			resp: user,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "signing up new user took too long")
		return
	case resp := <-userch:
		if resp.err != nil {
			status := 500
			switch err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		err := httpHelper.SendHTTPResponse(w, resp.resp)
		if err != nil {
			return
		}
	}
}

type userCreationResponse struct {
	userID int
	err    error
}

func (h *Handler) UserSignUp(w http.ResponseWriter, r *http.Request) {
	user, err := models.UnmarshalSignUpUser(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	userch := make(chan userCreationResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		userID, err := h.services.Authentication.SignUp(ctx, user.GetInternalUser())
		userch <- userCreationResponse{
			userID: int(userID),
			err:    err,
		}
	}()

	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "signing up new user took too long")
		return
	case resp := <-userch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.NotFoundError.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
		userResponse := models.UserCreationResponse{ID: resp.userID}

		err := httpHelper.SendHTTPResponse(w, userResponse)
		if err != nil {
			zlog.Log.Error(err, "could not send response")
			return
		}
	}
}
