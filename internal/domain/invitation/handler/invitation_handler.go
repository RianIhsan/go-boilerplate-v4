package handler

import (
	"encoding/json"
	"net/http"

	invdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/dto"
	invusecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/validator"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type InvitationHandler struct {
	invitationUsecase invusecase.InvitationUsecase
}

func NewInvitationHandler(invitationUsecase invusecase.InvitationUsecase) *InvitationHandler {
	return &InvitationHandler{invitationUsecase: invitationUsecase}
}

func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)

	var req invdto.CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.invitationUsecase.Create(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusCreated, "invitation created successfully", result)
}

func (h *InvitationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	id := chi.URLParam(r, "id")

	result, err := h.invitationUsecase.GetByID(r.Context(), id, userID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "invitation retrieved successfully", result)
}

func (h *InvitationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	pg := pagination.FromRequest(r)

	result, err := h.invitationUsecase.GetAll(r.Context(), userID, pg)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	meta := response.NewMeta(result.Page, result.Limit, result.TotalItems, result.TotalPages)
	response.SuccessList(w, "invitations retrieved successfully", result.Items, meta)
}

func (h *InvitationHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	id := chi.URLParam(r, "id")

	var req invdto.UpdateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.invitationUsecase.Update(r.Context(), id, userID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "invitation updated successfully", result)
}

func (h *InvitationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	id := chi.URLParam(r, "id")

	if err := h.invitationUsecase.Delete(r.Context(), id, userID); err != nil {
		response.Error(w, r, err)
		return
	}

	response.NoContent(w)
}

func (h *InvitationHandler) GetPublicBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	result, err := h.invitationUsecase.GetPublicBySlug(r.Context(), slug)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "invitation retrieved successfully", result)
}
