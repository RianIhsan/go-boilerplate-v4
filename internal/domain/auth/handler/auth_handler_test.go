package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/handler"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/golang/mock/gomock"
)

func TestAuthHandler_Register(t *testing.T) {
	mockResponse := &authdto.AuthResponse{
		AccessToken: "mock.jwt.token",
		User: authdto.UserResponse{
			ID:        "user-id",
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		},
	}

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(uc *mock.MockAuthUsecase)
		expectedStatus int
	}{
		{
			name: "success - 201",
			body: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			setupMock: func(uc *mock.MockAuthUsecase) {
				uc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(mockResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "error - invalid JSON body",
			body: "invalid json",
			setupMock: func(uc *mock.MockAuthUsecase) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "error - missing required fields",
			body: map[string]string{
				"name": "John",
			},
			setupMock:      func(uc *mock.MockAuthUsecase) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "error - email already exists",
			body: map[string]string{
				"name":     "John Doe",
				"email":    "existing@example.com",
				"password": "password123",
			},
			setupMock: func(uc *mock.MockAuthUsecase) {
				uc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, apperrors.ErrConflict)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "error - body exceeds max size",
			body: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
				"padding":  strings.Repeat("a", constants.MaxRequestBodyBytes+1),
			},
			setupMock:      func(uc *mock.MockAuthUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock.NewMockAuthUsecase(ctrl)
			tt.setupMock(mockUC)

			h := handler.NewAuthHandler(mockUC)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Register() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	mockResponse := &authdto.AuthResponse{
		AccessToken: "mock.jwt.token",
		User: authdto.UserResponse{
			ID:    "user-id",
			Email: "john@example.com",
		},
	}

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(uc *mock.MockAuthUsecase)
		expectedStatus int
	}{
		{
			name: "success - 200",
			body: map[string]string{
				"email":    "john@example.com",
				"password": "password123",
			},
			setupMock: func(uc *mock.MockAuthUsecase) {
				uc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(mockResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error - invalid credentials",
			body: map[string]string{
				"email":    "wrong@example.com",
				"password": "wrongpass",
			},
			setupMock: func(uc *mock.MockAuthUsecase) {
				uc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, apperrors.ErrInvalidCredential)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "error - missing fields",
			body:           map[string]string{},
			setupMock:      func(uc *mock.MockAuthUsecase) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "error - body exceeds max size",
			body: map[string]string{
				"email":    "john@example.com",
				"password": "password123",
				"padding":  strings.Repeat("a", constants.MaxRequestBodyBytes+1),
			},
			setupMock:      func(uc *mock.MockAuthUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock.NewMockAuthUsecase(ctrl)
			tt.setupMock(mockUC)

			h := handler.NewAuthHandler(mockUC)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Login() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}
