package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

// Init configures Sentry error reporting. An empty DSN disables the SDK
// entirely (local development default); production sets SENTRY_DSN.
func Init(dsn, env string) error {
	if dsn == "" {
		return nil
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		AttachStacktrace: true,
	})
}

// Flush waits for buffered events to reach Sentry (call on shutdown).
func Flush() {
	sentry.Flush(5 * time.Second)
}

// Middleware recovers panics, reports them to Sentry when enabled, and
// returns a generic 500 without leaking internals. It replaces chi's
// plain Recoverer so panics are observable in production.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
					hub.RecoverWithContext(r.Context(), rec)
				} else {
					sentry.CurrentHub().Recover(rec)
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":{"code":"INTERNAL","message":"Unexpected server error"}}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CaptureError reports a handled error with request context. It is a no-op
// when Sentry is disabled.
func CaptureError(r *http.Request, err error, extras map[string]interface{}) {
	if err == nil || sentry.CurrentHub().Client() == nil {
		return
	}
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetRequest(r)
		for k, v := range extras {
			scope.SetExtra(k, fmt.Sprint(v))
		}
		hub.CaptureException(err)
	})
}
