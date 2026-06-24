package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the auth domain's public routes onto r, guarded by
// rateLimitMiddleware since these endpoints have no auth gating of their own.
func RegisterRoutes(r chi.Router, h *AuthHandler, rateLimitMiddleware func(http.Handler) http.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Use(rateLimitMiddleware)

		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
}
