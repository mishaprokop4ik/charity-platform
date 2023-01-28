package handlers

import (
	httpHelper "Kurajj/pkg/http"
	"net/http"
	"strings"
)

func (h *Handler) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			httpHelper.SendErrorResponse(w, http.StatusUnauthorized, "empty auth header")
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			httpHelper.SendErrorResponse(w, http.StatusUnauthorized, "invalid auth header")
			return
		}
		_, err := h.services.Authentication.ParseToken(headerParts[1])
		if err != nil {
			httpHelper.SendErrorResponse(w, http.StatusUnauthorized, "invalid auth token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
