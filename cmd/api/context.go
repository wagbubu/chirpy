package main

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey = contextKey("user_id")

func (api *apiConfig) contextSetUserID(r *http.Request, user uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), userIDContextKey, user)
	return r.WithContext(ctx)
}

func (api *apiConfig) contextGetUserID(r *http.Request) uuid.UUID {
	user, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
	if !ok {
		panic("missing userID value in request context")
	}

	return user
}
