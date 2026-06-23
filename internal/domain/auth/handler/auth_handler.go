package handler

import (
	"encoding/json"
	"net/http"

	authdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/dto"
	authusecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/usecase"
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

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register payload"
// @Success 201 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
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

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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
