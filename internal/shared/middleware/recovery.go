package middleware

import (
	"net/http"

	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						zap.String("request_id", chimiddleware.GetReqID(r.Context())),
						zap.Any("panic", rec),
					)
					response.Error(w, r, apperrors.ErrInternalServer)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
