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

// UserSignIn signs in user into system
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
	user, err := models.UnmarshalSignInEntity(&r.Body)
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

	if user.Password == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("empty password"))
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
			case models.ErrNotFound.Error():
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

// UserSignUp creates a new user
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
	user, err := models.UnmarshalSignUpUser(&r.Body)
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

	if user.Password == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("empty password"))
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
			case models.ErrNotFound.Error():
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

// ConfirmEmail confirms that user's email is real and user has access to it
// @Summary      Updates user's status to 'activated'.
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
			case models.ErrNotFound.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}
	}
}

// RefreshTokens updates access token expiration date and returns access and refresh tokens
// @Summary      Update access token expiration date
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param request body models.RefreshTokenInput true "query params"
// @Success      200  {object}  models.TokensResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /auth/refresh-token [post]
func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	token, err := models.ParseRefresh(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	refreshch := make(chan refreshTokenResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		tokens, err := h.services.Authentication.RefreshTokens(ctx, token.RefreshToken)
		refreshch <- refreshTokenResponse{
			tokens: tokens,
			err:    err,
		}
	}()

	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "confirming email took too long")
		return
	case resp := <-refreshch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.ErrNotFound.Error():
				status = 404
			}
			httpHelper.SendErrorResponse(w, uint(status), resp.err.Error())
			return
		}

		tokenResponse := models.TokensResponse{
			AccessToken:  resp.tokens.Access,
			RefreshToken: resp.tokens.Refresh,
		}

		err := httpHelper.SendHTTPResponse(w, tokenResponse)
		if err != nil {
			zlog.Log.Error(err, "could not send response")
			return
		}
	}
}

// RefreshUserData returns user data
// @Summary      Return user data
// @Tags         User
// @Accept       json
// @Produce      json
// @Param request body models.RefreshTokenInput true "query params"
// @Success      201  {object}  models.SignedInUser
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/refresh-user-data [post]
func (h *Handler) RefreshUserData(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userch := make(chan userSignInResponse)
	userID := r.Context().Value("id")
	if userID == "" {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user transactionID isn't in context")
		return
	}
	token, err := models.ParseRefresh(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		member, err := h.services.Authentication.GetUserByRefreshToken(ctx, token.RefreshToken)
		userch <- userSignInResponse{
			resp: member,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "signing up new admin took too long")
		return
	case resp := <-userch:
		if resp.err != nil {
			status := 500
			switch resp.err.Error() {
			case models.ErrNotFound.Error():
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

// ReadNotifications changes notifications status to read=true
// @Summary      Changes notifications status to read=true
// @Tags         User
// @Accept       json
// @Produce      json
// @Param request body models.Ids true "query params"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/read-notifications [put]
func (h *Handler) ReadNotifications(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ids, err := models.ParseIds(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	errch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	go func() {
		err := h.services.TransactionNotification.Read(ctx, ids)
		errch <- errResponse{
			err: err,
		}
	}()

	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "reading notifications took too long")
		return
	case resp := <-errch:
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
