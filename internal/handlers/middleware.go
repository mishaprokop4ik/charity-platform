package handlers

import (
	httpHelper "Kurajj/pkg/http"
	zlog "Kurajj/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"strings"
)

const MemberIDContextKey = "id"

func (h *Handler) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		header := r.Header.Get("Authorization")
		if header == "" {
			httpHelper.SendErrorResponse(w, http.StatusUnauthorized, "empty auth header")
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, "invalid auth header")
			return
		}
		id, err := h.services.Authentication.ParseToken(headerParts[1])
		if err != nil {
			zlog.Log.Error(err, "incorrect input token")
			httpHelper.SendErrorResponse(w, http.StatusUnauthorized, "invalid auth token")
			return
		}
		ctx := r.Context()
		req := r.WithContext(context.WithValue(ctx, MemberIDContextKey, id))
		*r = *req
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) SetId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		header := r.Header.Get("Authorization")
		fmt.Println(header)
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			httpHelper.SendErrorResponse(w, http.StatusBadRequest, "invalid auth header")
			return
		}
		id, err := h.services.Authentication.ParseToken(headerParts[1])
		if err != nil {
			zlog.Log.Error(err, "incorrect input token")
			httpHelper.SendErrorResponse(w, http.StatusUnauthorized, "invalid auth token")
			return
		}
		ctx := r.Context()
		req := r.WithContext(context.WithValue(ctx, MemberIDContextKey, id))
		*r = *req
		next.ServeHTTP(w, r)
	})
}
