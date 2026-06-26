package handler

import (
	"encoding/json"
	"net/http"

	guestdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/dto"
	guestusecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/validator"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type GuestHandler struct {
	guestUsecase guestusecase.GuestUsecase
}

func NewGuestHandler(guestUsecase guestusecase.GuestUsecase) *GuestHandler {
	return &GuestHandler{guestUsecase: guestUsecase}
}

func (h *GuestHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	invitationID := chi.URLParam(r, "id")

	var req guestdto.CreateGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.guestUsecase.Create(r.Context(), userID, invitationID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusCreated, "guest created successfully", result)
}

func (h *GuestHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	invitationID := chi.URLParam(r, "id")
	pg := pagination.FromRequest(r)

	result, err := h.guestUsecase.GetAll(r.Context(), userID, invitationID, pg)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	meta := response.NewMeta(result.Page, result.Limit, result.TotalItems, result.TotalPages)
	response.SuccessList(w, "guests retrieved successfully", result.Items, meta)
}

func (h *GuestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	invitationID := chi.URLParam(r, "id")
	guestID := chi.URLParam(r, "guestId")

	if err := h.guestUsecase.Delete(r.Context(), userID, invitationID, guestID); err != nil {
		response.Error(w, r, err)
		return
	}

	response.NoContent(w)
}

func (h *GuestHandler) GetPublicByToken(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	token := chi.URLParam(r, "token")

	result, err := h.guestUsecase.GetPublicByToken(r.Context(), slug, token)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "guest retrieved successfully", result)
}
