package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *AuthHandler, rateLimitMiddleware func(http.Handler) http.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Use(rateLimitMiddleware)

		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
}
