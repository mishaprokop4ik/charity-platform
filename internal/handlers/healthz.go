package handlers

import "net/http"

func (h *Handler) handleHealthProbe(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
