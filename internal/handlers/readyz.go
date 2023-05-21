package handlers

import "net/http"

func (h *Handler) handleReadyzProbe(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
