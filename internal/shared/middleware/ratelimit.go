package middleware

import (
	"net/http"
	"time"

	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/go-chi/httprate"
)

// RateLimit returns a per-IP sliding-window rate limiter, intended for
// endpoints reachable without authentication (login/register) that would
// otherwise have no throttling against brute-force/credential-stuffing.
// Limit responses use the project's standard error envelope.
func RateLimit(requestLimit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(
		requestLimit,
		window,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			response.Error(w, r, apperrors.ErrTooManyRequests)
		}),
	)
}
