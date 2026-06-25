package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *FileMetaHandler, rateLimitMiddleware func(http.Handler) http.Handler) {
	r.Route("/files", func(r chi.Router) {
		r.Use(rateLimitMiddleware)

		r.Post("/metadata", h.ParseMetadata)
	})
}
