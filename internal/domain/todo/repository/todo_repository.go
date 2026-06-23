package repository

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/entity"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=todo_repository.go -destination=../../../mock/mock_todo_repository.go -package=mock

type TodoRepository interface {
	Create(ctx context.Context, todo *entity.Todo) error
	FindByID(ctx context.Context, id, userID string) (*entity.Todo, error)
	FindAllByUserID(ctx context.Context, userID string, pg pagination.Pagination) ([]*entity.Todo, int64, error)
	Update(ctx context.Context, todo *entity.Todo) error
	Delete(ctx context.Context, id, userID string) error
}
