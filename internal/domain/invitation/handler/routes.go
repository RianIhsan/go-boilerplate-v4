package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *InvitationHandler, authMiddleware, publicRateLimit func(http.Handler) http.Handler) {
	r.Route("/invitations", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", h.Create)
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	r.Route("/public/invitations", func(r chi.Router) {
		r.Use(publicRateLimit)

		r.Get("/{slug}", h.GetPublicBySlug)
	})
}
