package handlers

import (
	"Kurajj/internal/models"
	httpHelper "Kurajj/pkg/http"
	"bytes"
	"fmt"
	"io"
	"net/http"
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

	fmt.Println(header.Filename)

	var fileBytes bytes.Buffer
	_, err = io.Copy(&fileBytes, file)
	if err != nil {
		return
	}

	filePath, err := h.services.Upload(r.Context(), header.Filename, &fileBytes)
	if err != nil {
		return
	}

	fmt.Println(filePath)

	resp := models.File{Path: filePath}

	err = httpHelper.SendHTTPResponse(w, resp)
	if err != nil {
		return
	}
}
