package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	tododto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/entity"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/golang/mock/gomock"
)

var (
	mockUserID = "user-id-1"
	mockTodoID = "todo-id-1"
	mockTodo   = &entity.Todo{
		ID:          mockTodoID,
		UserID:      mockUserID,
		Title:       "Buy groceries",
		Description: "Milk, Eggs, Bread",
		Status:      entity.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
)

func TestTodoUsecase_Create(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		req         *tododto.CreateTodoRequest
		setupMock   func(repo *mock.MockTodoRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "success - todo created",
			userID: mockUserID,
			req: &tododto.CreateTodoRequest{
				Title:       "Buy groceries",
				Description: "Milk, Eggs, Bread",
			},
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "error - repository fails",
			userID: mockUserID,
			req: &tododto.CreateTodoRequest{
				Title: "Failed todo",
			},
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockTodoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTodoUsecase(mockRepo)
			got, err := uc.Create(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Create() expected error but got nil")
				}
				if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("Create() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
			}
			if got == nil || got.Title != tt.req.Title {
				t.Errorf("Create() title mismatch, got %v", got)
			}
		})
	}
}

func TestTodoUsecase_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		todoID      string
		userID      string
		setupMock   func(repo *mock.MockTodoRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "success - todo found",
			todoID: mockTodoID,
			userID: mockUserID,
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockTodoID, mockUserID).Return(mockTodo, nil)
			},
			wantErr: false,
		},
		{
			name:   "error - todo not found",
			todoID: "nonexistent-id",
			userID: mockUserID,
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "nonexistent-id", mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			expectedErr: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockTodoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTodoUsecase(mockRepo)
			got, err := uc.GetByID(context.Background(), tt.todoID, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetByID() unexpected error = %v", err)
			}
			if got == nil || got.ID != tt.todoID {
				t.Errorf("GetByID() got wrong todo")
			}
		})
	}
}

func TestTodoUsecase_GetAll(t *testing.T) {
	pg := pagination.Pagination{Page: 1, Limit: 10, Offset: 0}

	tests := []struct {
		name       string
		userID     string
		pagination pagination.Pagination
		setupMock  func(repo *mock.MockTodoRepository)
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "success - returns list",
			userID:     mockUserID,
			pagination: pg,
			setupMock: func(repo *mock.MockTodoRepository) {
				todos := []*entity.Todo{mockTodo, {
					ID: "todo-id-2", UserID: mockUserID, Title: "Second todo",
					Status: entity.StatusDone, CreatedAt: time.Now(), UpdatedAt: time.Now(),
				}}
				repo.EXPECT().FindAllByUserID(gomock.Any(), mockUserID, pg).Return(todos, int64(2), nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:       "success - empty list",
			userID:     "no-todos-user",
			pagination: pg,
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindAllByUserID(gomock.Any(), "no-todos-user", pg).Return([]*entity.Todo{}, int64(0), nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:       "error - repository fails",
			userID:     mockUserID,
			pagination: pg,
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindAllByUserID(gomock.Any(), mockUserID, pg).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockTodoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTodoUsecase(mockRepo)
			got, err := uc.GetAll(context.Background(), tt.userID, tt.pagination)

			if tt.wantErr {
				if err == nil {
					t.Error("GetAll() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetAll() unexpected error = %v", err)
			}
			if len(got.Items) != tt.wantCount {
				t.Errorf("GetAll() items count = %d, want %d", len(got.Items), tt.wantCount)
			}
		})
	}
}

func TestTodoUsecase_Update(t *testing.T) {
	tests := []struct {
		name        string
		todoID      string
		userID      string
		req         *tododto.UpdateTodoRequest
		setupMock   func(repo *mock.MockTodoRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "success - update title and status",
			todoID: mockTodoID,
			userID: mockUserID,
			req:    &tododto.UpdateTodoRequest{Title: "Updated Title", Status: "done"},
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockTodoID, mockUserID).Return(mockTodo, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "error - todo not found",
			todoID: "ghost-id",
			userID: mockUserID,
			req:    &tododto.UpdateTodoRequest{Title: "New Title"},
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "ghost-id", mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			expectedErr: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockTodoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTodoUsecase(mockRepo)
			_, err := uc.Update(context.Background(), tt.todoID, tt.userID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Update() unexpected error = %v", err)
			}
		})
	}
}

func TestTodoUsecase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		todoID    string
		userID    string
		setupMock func(repo *mock.MockTodoRepository)
		wantErr   bool
	}{
		{
			name:   "success - deleted",
			todoID: mockTodoID,
			userID: mockUserID,
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockTodoID, mockUserID).Return(mockTodo, nil)
				repo.EXPECT().Delete(gomock.Any(), mockTodoID, mockUserID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "error - not found",
			todoID: "ghost-id",
			userID: mockUserID,
			setupMock: func(repo *mock.MockTodoRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "ghost-id", mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockTodoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTodoUsecase(mockRepo)
			err := uc.Delete(context.Background(), tt.todoID, tt.userID)

			if tt.wantErr && err == nil {
				t.Error("Delete() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}
		})
	}
}
