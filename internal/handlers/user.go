package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type userSignInResponse struct {
	resp models.SignedInUser
	err  error
}

// UserSignIn godoc
// @Summary      Signs In a user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param request body models.SignInEntity true "query params"
// @Success      201  {object}  models.SignedInUser
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /auth/sign-in [post]
func (h *Handler) UserSignIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user, err := models.UnmarshalSignInEntity(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err := user.Email.Validate()
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !ok {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("email: %s is incorrect", user.Email))
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
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "signing in user took too long")
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
		err := httpHelper.SendHTTPResponse(w, resp.resp)
		if err != nil {
			return
		}
	}
}

type idResponse struct {
	err error
	id  int
}

type errResponse struct {
	err error
}

// UserSignUp godoc
// @Summary      Signs Up new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param request body models.SignUpUser true "query params"
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /auth/sign-up [post]
func (h *Handler) UserSignUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user, err := models.UnmarshalSignUpUser(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err := user.Email.Validate()
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !ok {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("email: %s is incorrect", user.Email))
		return
	}

	ok = user.Telephone.Validate()

	if !ok {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("phone number: %s is incorrect", user.Telephone))
		return
	}

	user.Telephone = user.Telephone.GetDefaultTelephoneNumber()
	userch := make(chan idResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		userID, err := h.services.Authentication.SignUp(ctx, user.GetInternalUser())
		userch <- idResponse{
			id:  int(userID),
			err: err,
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
		userResponse := models.CreationResponse{ID: resp.id}

		err := httpHelper.SendHTTPResponse(w, userResponse)
		if err != nil {
			zlog.Log.Error(err, "could not send response")
			return
		}
	}
}

// ConfirmEmail godoc
// @Summary      Updates user's status to activated.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        email   path string  true  "Email"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /auth/confirm/{email} [post]
func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	email, ok := mux.Vars(r)["email"]
	validateEmail, _ := models.Email(email).Validate()
	if !ok || !validateEmail {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "there is no email in params")
		return
	}
	userch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		err := h.services.Authentication.ConfirmEmail(ctx, email)
		userch <- errResponse{
			err: err,
		}
	}()

	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "confirming email took too long")
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
	}
}
