package repository

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/entity"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=invitation_repository.go -destination=../../../mock/mock_invitation_repository.go -package=mock

type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.Invitation) error
	FindByID(ctx context.Context, id, userID string) (*entity.Invitation, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Invitation, error)
	FindAllByUserID(ctx context.Context, userID string, pg pagination.Pagination) ([]*entity.Invitation, int64, error)
	Update(ctx context.Context, invitation *entity.Invitation) error
	Delete(ctx context.Context, id, userID string) error
}
