package sentry

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

type SentryConfig struct {
	Enabled bool   `cfg:"SENTRY_ENABLED" default:"true"`
	DSN     string `cfg:"SENTRY_DSN"`
}

func InitSentry(environment string, cfg *SentryConfig) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Environment: environment,
	})
	if err != nil {
		return err
	}

	sentry.CaptureMessage("It works!")
	sentry.Flush(time.Second * 5)

	return nil
}

func SentryRecover(ctx context.Context, err any) {
	hub := sentry.GetHubFromContext(ctx)
	if hub != nil {
		hub.RecoverWithContext(ctx, err)
	} else {
		sentry.CurrentHub().RecoverWithContext(ctx, err)
	}
}

func SentryCaptureException(ctx context.Context, err error, tags map[string]string) {
	hub := sentry.GetHubFromContext(ctx)
	if hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for key, value := range tags {
				scope.SetTag(key, value)
			}
			hub.CaptureException(err)
		})
	} else {
		sentry.CurrentHub().CaptureException(err)
	}
}

func SentryMiddleware(handler http.Handler) http.Handler {
	mw := sentryhttp.New(sentryhttp.Options{})
	return mw.Handle(handler)
}
