package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	"context"
	"net/http"
	"time"
)

// UpsertUserSearch   deletes old user search values and create new by input
// @Summary      Deletes old user search values and create new by input
// @Tags         Tag
// @Accept       json
// @Param request body models.MemberSearchValuesRequestCreate true "query params"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/tags/user-search [post]
func (h *Handler) UpsertUserSearch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	searchValue, err := models.UnmarshalSearchValuesGroupCreateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := r.Context().Value("id")
	if userID == "" || userID == nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, "user transactionID isn't in context")
		return
	}

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		respError := h.services.UserSearchValue.UpsertValues(ctx, userID.(uint), searchValue.Internal())

		eventch <- errResponse{
			err: respError,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "upsert user search tags took too long")
		return
	case resp := <-eventch:
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
