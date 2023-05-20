package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// handleUploadFile uploads file
// @Param request body models.CommentCreateRequest true "query params"
// @Tags         Help Event
// @Summary      CreateNotification new comment in help event
// @SearchValuesResponse         Help Event
// @Accept       json
// @Success      201  {object}  models.CreationResponse
// @Failure      401  {object}  models.ErrResponse
// @Failure      403  {object}  models.ErrResponse
// @Failure      404  {object}  models.ErrResponse
// @Failure      408  {object}  models.ErrResponse
// @Failure      500  {object}  models.ErrResponse
// @Router       /api/events/help/comment [post]
func (h *Handler) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return
	}
	defer file.Close()

	var fileBytes bytes.Buffer
	_, err = io.Copy(&fileBytes, file)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	filePath, err := h.services.Upload(r.Context(), header.Filename, &fileBytes)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := models.File{Path: filePath}

	err = httpHelper.SendHTTPResponse(w, resp)
	if err != nil {
		return
	}
}

func (h *Handler) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	imagePath, err := models.NewFilePath(&r.Body)
	if err != nil {
		httpHelper.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	imagePathes := strings.Split(imagePath.Path, "s3.amazonaws.com/")
	imageName := imagePathes[len(imagePathes)-1]
	eventch := make(chan errResponse)
	defer close(eventch)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {

		err := h.services.Delete(r.Context(), imageName)
		if err != nil {
			httpHelper.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		eventch <- errResponse{
			err: err,
		}
	}()
	select {
	case <-ctx.Done():
		httpHelper.SendErrorResponse(w, http.StatusRequestTimeout, "deleting proposal event took too long")
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
