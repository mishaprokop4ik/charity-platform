package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	"context"
	"fmt"
	"net/http"
	"time"
)

// AdminSignIn godoc
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
// @Router       /auth/sign-in-admin [post]
func (h *Handler) AdminSignIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	admin, err := models.UnmarshalSignInEntity(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err := admin.Email.Validate()
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !ok {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("email: %s is incorrect", admin.Email))
		return
	}

	userch := make(chan userSignInResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		user, err := h.services.Authentication.SignIn(ctx,
			models.User{
				Email:    string(admin.Email),
				Password: admin.Password,
				IsAdmin:  true,
			})
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

// CreateNewAdmin godoc
// @Summary      Create a new admin
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param request body models.AdminCreation true "query params"
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/admin/create [post]
func (h *Handler) CreateNewAdmin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	admin, err := models.UnmarshalCreateAdmin(r)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err := admin.Email.Validate()
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !ok {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("email: %s is incorrect", admin.Email))
		return
	}

	newAdmin := admin.CreateUser()

	userch := make(chan idResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		id, err := h.services.Admin.CreateAdmin(ctx, newAdmin)
		userch <- idResponse{
			id:  int(id),
			err: err,
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
		adminResponse := models.CreationResponse{ID: resp.id}
		err := httpHelper.SendHTTPResponse(w, adminResponse)
		if err != nil {
			return
		}
	}
}
