package handler

import (
	"encoding/json"
	"net/http"

	rsvpdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/dto"
	rsvpusecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/validator"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type RSVPHandler struct {
	rsvpUsecase rsvpusecase.RSVPUsecase
}

func NewRSVPHandler(rsvpUsecase rsvpusecase.RSVPUsecase) *RSVPHandler {
	return &RSVPHandler{rsvpUsecase: rsvpUsecase}
}

func (h *RSVPHandler) Submit(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxRequestBodyBytes)

	var req rsvpdto.SubmitRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.rsvpUsecase.Submit(r.Context(), slug, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusCreated, "rsvp submitted successfully", result)
}

func (h *RSVPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	invitationID := chi.URLParam(r, "id")
	pg := pagination.FromRequest(r)

	result, err := h.rsvpUsecase.GetAll(r.Context(), userID, invitationID, pg)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	meta := response.NewMeta(result.Page, result.Limit, result.TotalItems, result.TotalPages)
	response.SuccessList(w, "rsvps retrieved successfully", result.Items, meta)
}
