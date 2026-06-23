package usecase

import (
	"context"
	"time"

	tododto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/entity"
	todorepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/google/uuid"
)

type todoUsecaseImpl struct {
	todoRepo todorepo.TodoRepository
}

func NewTodoUsecase(todoRepo todorepo.TodoRepository) TodoUsecase {
	return &todoUsecaseImpl{todoRepo: todoRepo}
}

func (u *todoUsecaseImpl) Create(ctx context.Context, userID string, req *tododto.CreateTodoRequest) (*tododto.TodoResponse, error) {
	todo := &entity.Todo{
		ID:          uuid.NewString(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.todoRepo.Create(ctx, todo); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return toResponse(todo), nil
}

func (u *todoUsecaseImpl) GetByID(ctx context.Context, id, userID string) (*tododto.TodoResponse, error) {
	todo, err := u.todoRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, apperrors.TodoNotFound(id)
	}
	return toResponse(todo), nil
}

func (u *todoUsecaseImpl) GetAll(ctx context.Context, userID string, pg pagination.Pagination) (*tododto.TodoListResponse, error) {
	todos, total, err := u.todoRepo.FindAllByUserID(ctx, userID, pg)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	items := make([]*tododto.TodoResponse, 0, len(todos))
	for _, t := range todos {
		items = append(items, toResponse(t))
	}

	return &tododto.TodoListResponse{
		Items:      items,
		TotalItems: total,
		Page:       pg.Page,
		Limit:      pg.Limit,
		TotalPages: pagination.TotalPages(total, pg.Limit),
	}, nil
}

func (u *todoUsecaseImpl) Update(ctx context.Context, id, userID string, req *tododto.UpdateTodoRequest) (*tododto.TodoResponse, error) {
	todo, err := u.todoRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, apperrors.TodoNotFound(id)
	}

	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Description != "" {
		todo.Description = req.Description
	}
	if req.Status != "" {
		todo.Status = entity.TodoStatus(req.Status)
	}
	todo.UpdatedAt = time.Now()

	if err := u.todoRepo.Update(ctx, todo); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return toResponse(todo), nil
}

func (u *todoUsecaseImpl) Delete(ctx context.Context, id, userID string) error {
	_, err := u.todoRepo.FindByID(ctx, id, userID)
	if err != nil {
		return apperrors.TodoNotFound(id)
	}

	if err := u.todoRepo.Delete(ctx, id, userID); err != nil {
		return apperrors.Wrap(apperrors.ErrInternalServer, err)
	}
	return nil
}

func toResponse(t *entity.Todo) *tododto.TodoResponse {
	return &tododto.TodoResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
