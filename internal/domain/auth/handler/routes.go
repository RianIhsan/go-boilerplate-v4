package handler

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts the auth domain's public routes onto r.
func RegisterRoutes(r chi.Router, h *AuthHandler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
}
