package handler

import (
	"encoding/json"
	"net/http"

	tododto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/dto"
	todousecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/validator"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type TodoHandler struct {
	todoUsecase todousecase.TodoUsecase
}

func NewTodoHandler(todoUsecase todousecase.TodoUsecase) *TodoHandler {
	return &TodoHandler{todoUsecase: todoUsecase}
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)

	var req tododto.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.todoUsecase.Create(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusCreated, "todo created successfully", result)
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	id := chi.URLParam(r, "id")

	result, err := h.todoUsecase.GetByID(r.Context(), id, userID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "todo retrieved successfully", result)
}

func (h *TodoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	pg := pagination.FromRequest(r)

	result, err := h.todoUsecase.GetAll(r.Context(), userID, pg)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	meta := response.NewMeta(result.Page, result.Limit, result.TotalItems, result.TotalPages)
	response.SuccessList(w, "todos retrieved successfully", result.Items, meta)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	id := chi.URLParam(r, "id")

	var req tododto.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.todoUsecase.Update(r.Context(), id, userID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "todo updated successfully", result)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.ContextKeyUserID).(string)
	id := chi.URLParam(r, "id")

	if err := h.todoUsecase.Delete(r.Context(), id, userID); err != nil {
		response.Error(w, r, err)
		return
	}

	response.NoContent(w)
}
