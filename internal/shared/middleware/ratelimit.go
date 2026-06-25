package middleware

import (
	"net/http"
	"time"

	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/go-chi/httprate"
)

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
