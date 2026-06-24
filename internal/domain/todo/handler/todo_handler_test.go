package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func buildBody(body interface{}) io.Reader {
	switch v := body.(type) {
	case string:
		return bytes.NewBufferString(v)
	case nil:
		return nil
	default:
		b, _ := json.Marshal(v)
		return bytes.NewBuffer(b)
	}
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
			name:           "error - invalid JSON",
			body:           "not json at all {{{",
			setupMock:      func(uc *mock.MockTodoUsecase) {},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPost, "/todos", buildBody(tt.body))
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
			req = withChiURLParam(req, "id", tt.todoID)
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

func TestTodoHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		body           interface{}
		setupMock      func(uc *mock.MockTodoUsecase)
		expectedStatus int
	}{
		{
			name:   "success - 200",
			todoID: "todo-id-1",
			body:   map[string]string{"title": "Updated title", "status": "done"},
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().Update(gomock.Any(), "todo-id-1", "user-id-1", gomock.Any()).Return(mockTodoResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - invalid JSON",
			todoID:         "todo-id-1",
			body:           "not json at all {{{",
			setupMock:      func(uc *mock.MockTodoUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - validation fails",
			todoID:         "todo-id-1",
			body:           map[string]string{"status": "invalid_status"},
			setupMock:      func(uc *mock.MockTodoUsecase) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "error - usecase fails",
			todoID: "todo-id-1",
			body:   map[string]string{"title": "Update"},
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().Update(gomock.Any(), "todo-id-1", "user-id-1", gomock.Any()).Return(nil, apperrors.ErrNotFound)
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

			req := httptest.NewRequest(http.MethodPut, "/todos/"+tt.todoID, buildBody(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = contextWithUser(req, "user-id-1")
			req = withChiURLParam(req, "id", tt.todoID)
			rr := httptest.NewRecorder()

			h.Update(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Update() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestTodoHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		setupMock      func(uc *mock.MockTodoUsecase)
		expectedStatus int
	}{
		{
			name:   "success - 204",
			todoID: "todo-id-1",
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().Delete(gomock.Any(), "todo-id-1", "user-id-1").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "error - not found",
			todoID: "ghost-id",
			setupMock: func(uc *mock.MockTodoUsecase) {
				uc.EXPECT().Delete(gomock.Any(), "ghost-id", "user-id-1").Return(apperrors.ErrNotFound)
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

			req := httptest.NewRequest(http.MethodDelete, "/todos/"+tt.todoID, nil)
			req = contextWithUser(req, "user-id-1")
			req = withChiURLParam(req, "id", tt.todoID)
			rr := httptest.NewRecorder()

			h.Delete(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Delete() status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestRegisterRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockTodoUsecase(ctrl)
	mockList := &tododto.TodoListResponse{
		Items:      []*tododto.TodoResponse{mockTodoResponse},
		TotalItems: 1,
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}

	mockUC.EXPECT().Create(gomock.Any(), "user-id-1", gomock.Any()).Return(mockTodoResponse, nil)
	mockUC.EXPECT().GetAll(gomock.Any(), "user-id-1", gomock.Any()).Return(mockList, nil)
	mockUC.EXPECT().GetByID(gomock.Any(), "todo-id-1", "user-id-1").Return(mockTodoResponse, nil)
	mockUC.EXPECT().Update(gomock.Any(), "todo-id-1", "user-id-1", gomock.Any()).Return(mockTodoResponse, nil)
	mockUC.EXPECT().Delete(gomock.Any(), "todo-id-1", "user-id-1").Return(nil)

	h := handler.NewTodoHandler(mockUC)

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, "user-id-1")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	r := chi.NewRouter()
	handler.RegisterRoutes(r, h, authMiddleware)

	routes := []struct {
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{http.MethodPost, "/todos/", map[string]string{"title": "Test todo"}, http.StatusCreated},
		{http.MethodGet, "/todos/", nil, http.StatusOK},
		{http.MethodGet, "/todos/todo-id-1", nil, http.StatusOK},
		{http.MethodPut, "/todos/todo-id-1", map[string]string{"title": "Updated"}, http.StatusOK},
		{http.MethodDelete, "/todos/todo-id-1", nil, http.StatusNoContent},
	}

	for _, tt := range routes {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, buildBody(tt.body))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("%s %s: status = %d, want %d, body: %s",
					tt.method, tt.path, rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}
