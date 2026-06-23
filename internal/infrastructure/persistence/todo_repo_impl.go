package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/entity"
	todorepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

type todoRepositoryImpl struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) todorepo.TodoRepository {
	return &todoRepositoryImpl{db: db}
}

func (r *todoRepositoryImpl) Create(ctx context.Context, todo *entity.Todo) error {
	query := `
		INSERT INTO todos (id, user_id, title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		todo.ID, todo.UserID, todo.Title, todo.Description, todo.Status, todo.CreatedAt, todo.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("todoRepository.Create: %w", err)
	}
	return nil
}

func (r *todoRepositoryImpl) FindByID(ctx context.Context, id, userID string) (*entity.Todo, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM todos
		WHERE id = $1 AND user_id = $2
	`
	todo := &entity.Todo{}
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Status, &todo.CreatedAt, &todo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("todoRepository.FindByID: %w", err)
	}
	return todo, nil
}

func (r *todoRepositoryImpl) FindAllByUserID(ctx context.Context, userID string, pg pagination.Pagination) ([]*entity.Todo, int64, error) {
	countQuery := `SELECT COUNT(*) FROM todos WHERE user_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("todoRepository.FindAllByUserID count: %w", err)
	}

	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pg.Limit, pg.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("todoRepository.FindAllByUserID: %w", err)
	}
	defer rows.Close()

	var todos []*entity.Todo
	for rows.Next() {
		todo := &entity.Todo{}
		if err := rows.Scan(
			&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Status, &todo.CreatedAt, &todo.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("todoRepository.FindAllByUserID scan: %w", err)
		}
		todos = append(todos, todo)
	}

	return todos, total, nil
}

func (r *todoRepositoryImpl) Update(ctx context.Context, todo *entity.Todo) error {
	query := `
		UPDATE todos
		SET title = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`
	result, err := r.db.ExecContext(ctx, query,
		todo.Title, todo.Description, todo.Status, todo.UpdatedAt, todo.ID, todo.UserID,
	)
	if err != nil {
		return fmt.Errorf("todoRepository.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *todoRepositoryImpl) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM todos WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("todoRepository.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
