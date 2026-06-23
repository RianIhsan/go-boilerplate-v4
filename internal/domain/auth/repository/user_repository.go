package repository

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/entity"
)

//go:generate mockgen -source=user_repository.go -destination=../../../mock/mock_user_repository.go -package=mock

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
}
