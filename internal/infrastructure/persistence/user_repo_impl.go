package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/entity"
	authrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/lib/pq"
)

const pqUniqueViolation = "23505"

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) authrepo.UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Name, user.Email, user.Password, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pqUniqueViolation {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("userRepository.Create: %w", err)
	}
	return nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepository.FindByEmail: %w", err)
	}
	return user, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return user, nil
}
