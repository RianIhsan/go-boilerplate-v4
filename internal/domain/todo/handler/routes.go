package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *TodoHandler, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/todos", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", h.Create)
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}
