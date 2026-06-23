package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/jwt"
)

func Auth(jwtSvc jwt.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, r, apperrors.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, r, apperrors.ErrInvalidToken)
				return
			}

			claims, err := jwtSvc.ValidateToken(parts[1])
			if err != nil {
				response.Error(w, r, apperrors.ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, constants.ContextKeyEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
