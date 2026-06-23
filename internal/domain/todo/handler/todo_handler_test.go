package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tododto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/handler"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
)

var mockTodoResponse = &tododto.TodoResponse{
	ID:        "todo-id-1",
	UserID:    "user-id-1",
	Title:     "Buy groceries",
	Status:    "pending",
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func contextWithUser(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, userID)
	return r.WithContext(ctx)
}

func TestTodoHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(uc *mock.MockTodoUsecase)
		expectedStatus int
	}{
		{
			name: "success - 201",
			body: map[string]string{"title": "Buy groceries", "description": "Milk"},
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().Create(gomock.Any(), "user-id-1", gomock.Any()).Return(mockTodoResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "error - missing title",
			body:           map[string]string{"description": "no title"},
			setupMock:      func(uc *mock.MockTodoUsecase) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "error - usecase fails",
			body: map[string]string{"title": "Test"},
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().Create(gomock.Any(), "user-id-1", gomock.Any()).Return(nil, apperrors.ErrInternalServer)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock.NewMockTodoUsecase(ctrl)
			tt.setupMock(mockUC)
			h := handler.NewTodoHandler(mockUC)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = contextWithUser(req, "user-id-1")
			rr := httptest.NewRecorder()

			h.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Create() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestTodoHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		setupMock      func(uc *mock.MockTodoUsecase)
		expectedStatus int
	}{
		{
			name:   "success - 200",
			todoID: "todo-id-1",
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().GetByID(gomock.Any(), "todo-id-1", "user-id-1").Return(mockTodoResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "error - not found",
			todoID: "ghost-id",
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().GetByID(gomock.Any(), "ghost-id", "user-id-1").Return(nil, apperrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock.NewMockTodoUsecase(ctrl)
			tt.setupMock(mockUC)
			h := handler.NewTodoHandler(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/todos/"+tt.todoID, nil)
			req = contextWithUser(req, "user-id-1")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.todoID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.GetByID(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("GetByID() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestTodoHandler_GetAll(t *testing.T) {
	mockList := &tododto.TodoListResponse{
		Items:      []*tododto.TodoResponse{mockTodoResponse},
		TotalItems: 1,
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}

	tests := []struct {
		name           string
		setupMock      func(uc *mock.MockTodoUsecase)
		expectedStatus int
	}{
		{
			name: "success - 200",
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().GetAll(gomock.Any(), "user-id-1", gomock.Any()).Return(mockList, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error - internal server",
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().GetAll(gomock.Any(), "user-id-1", gomock.Any()).Return(nil, apperrors.ErrInternalServer)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock.NewMockTodoUsecase(ctrl)
			tt.setupMock(mockUC)
			h := handler.NewTodoHandler(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/todos?page=1&limit=10", nil)
			req = contextWithUser(req, "user-id-1")
			_ = pagination.FromRequest(req)
			rr := httptest.NewRecorder()

			h.GetAll(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("GetAll() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}
