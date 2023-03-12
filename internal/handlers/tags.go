package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

// UpsertTags   deletes all previous tags and their values and creates new by input
// @Summary      Delete all previous tags and their values and create new by input
// @Tags         Tag
// @Accept       json
// @Param request body models.TagGroupRequestCreate true "query params"
// @Success      200
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/tags/upsert [post]
func (h *Handler) UpsertTags(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tags, err := models.UnmarshalTagGroupCreateRequest(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	eventch := make(chan errResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		respError := h.services.Tag.UpsertTags(ctx, tags.EventID, tags.EventType, tags.Internal())

		eventch <- errResponse{
			err: respError,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting all transactions for proposal event took too long")
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

func (h *Handler) GetProposalEventTags(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id, ok := mux.Vars(r)["id"]
	parsedID, err := strconv.Atoi(id)
	if !ok || err != nil {
		response := "there is no id for getting in URL"
		if err != nil {
			response = err.Error()
		}
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	eventch := make(chan getTagsResponse)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		tags, respError := h.services.Tag.GetTagsByEvent(ctx, uint(parsedID), models.ProposalEventType)

		eventch <- getTagsResponse{
			tags: tags,
			err:  respError,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting all transactions for proposal event took too long")
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
		tagsResponse := models.Tags{
			Tags: make([]models.TagsResponse, len(resp.tags)),
		}
		for i, t := range resp.tags {
			tagValues := make([]models.TagValueResponse, len(resp.tags[i].Values))
			for j, tagValue := range t.Values {
				tagValues[j] = models.TagValueResponse{
					ID:    tagValue.ID,
					Value: tagValue.Value,
				}
			}
			tagsResponse.Tags[i] = models.TagsResponse{
				ID:        t.ID,
				Title:     t.Title,
				EventID:   t.EventID,
				EventType: t.EventType,
				Values:    tagValues,
			}
		}
		err = httpHelper.SendHTTPResponse(w, tagsResponse)
		if err != nil {
			zlog.Log.Error(err, "could not send proposal events")
		}
	}
}
