package handler

import (
	"encoding/json"
	"net/http"

	authdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/dto"
	authusecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/validator"
)

type AuthHandler struct {
	authUsecase authusecase.AuthUsecase
}

func NewAuthHandler(authUsecase authusecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxRequestBodyBytes)

	var req authdto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.authUsecase.Register(r.Context(), &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusCreated, "user registered successfully", result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxRequestBodyBytes)

	var req authdto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	if fields := validator.Validate(req); fields != nil {
		response.ValidationError(w, r, fields)
		return
	}

	result, err := h.authUsecase.Login(r.Context(), &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "login successful", result)
}
