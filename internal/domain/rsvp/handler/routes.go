package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *RSVPHandler, authMiddleware, publicRateLimit func(http.Handler) http.Handler) {
	r.Route("/invitations/{id}/rsvps", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", h.GetAll)
	})

	r.Route("/public/invitations/{slug}/rsvp", func(r chi.Router) {
		r.Use(publicRateLimit)

		r.Post("/", h.Submit)
	})
}
