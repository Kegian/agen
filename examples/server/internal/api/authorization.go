package api

import (
	"context"

	"github.com/Kegian/agen/examples/server/internal/generated/oapi"
)

type ContextKey string

var (
	UserIDKey ContextKey = "userID"
)

type SecurityHandler struct{}

func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, _ string, t oapi.BearerAuth) (context.Context, error) {
	if t.Token == "" {
		return ctx, nil
	}

	// TODO: Getting user id here
	userID := t.Token

	return context.WithValue(ctx, UserIDKey, userID), nil
}
