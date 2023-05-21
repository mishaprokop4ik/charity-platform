package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
	"context"
	"net/http"
	"time"
)

// handleGetGlobalStatistics takes statistics of all events from time.Now() - 28
// @Summary      Take statistics of events from current date - 28 to current date
// @SearchValuesResponse         Proposal Event
// @Success      200  {object}  models.GlobalStatistics
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /open-api/statistics [get]
func (h *Handler) handleGetGlobalStatistics(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	eventch := make(chan getGlobalStatistics)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		statistics, err := h.services.GetGlobalStatistics(ctx, 28)

		eventch <- getGlobalStatistics{
			statistics: statistics,
			err:        err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "getting proposal event took too long")
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
		err := httpHelper.SendHTTPResponse(w, resp.statistics)
		if err != nil {
			zlog.Log.Error(err, "got an error")
		}
	}
}
