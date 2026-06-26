package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *GuestHandler, authMiddleware, publicRateLimit func(http.Handler) http.Handler) {
	r.Route("/invitations/{id}/guests", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", h.Create)
		r.Get("/", h.GetAll)
		r.Delete("/{guestId}", h.Delete)
	})

	r.Route("/public/invitations/{slug}/guests", func(r chi.Router) {
		r.Use(publicRateLimit)

		r.Get("/{token}", h.GetPublicByToken)
	})
}
