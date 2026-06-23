package usecase

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/dto"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=todo_usecase.go -destination=../../../mock/mock_todo_usecase.go -package=mock

type TodoUsecase interface {
	Create(ctx context.Context, userID string, req *dto.CreateTodoRequest) (*dto.TodoResponse, error)
	GetByID(ctx context.Context, id, userID string) (*dto.TodoResponse, error)
	GetAll(ctx context.Context, userID string, pg pagination.Pagination) (*dto.TodoListResponse, error)
	Update(ctx context.Context, id, userID string, req *dto.UpdateTodoRequest) (*dto.TodoResponse, error)
	Delete(ctx context.Context, id, userID string) error
}
