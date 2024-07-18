package agen

import (
	"context"
	"net/http"

	"github.com/Kegian/agen/sentry"
)

func InitSentry(environment string, cfg *sentry.SentryConfig) error {
	return sentry.InitSentry(environment, cfg)
}

func SentryRecover(ctx context.Context, err any) {
	sentry.SentryRecover(ctx, err)
}

func SentryCaptureException(ctx context.Context, err error, tags map[string]string) {
	sentry.SentryCaptureException(ctx, err, tags)
}

func SentryMiddleware(handler http.Handler) http.Handler {
	return sentry.SentryMiddleware(handler)
}
