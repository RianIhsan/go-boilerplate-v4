package repository

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/entity"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=guest_repository.go -destination=../../../mock/mock_guest_repository.go -package=mock

type GuestRepository interface {
	Create(ctx context.Context, guest *entity.Guest) error
	FindByID(ctx context.Context, id, invitationID string) (*entity.Guest, error)
	FindByToken(ctx context.Context, token string) (*entity.Guest, error)
	FindAllByInvitationID(ctx context.Context, invitationID string, pg pagination.Pagination) ([]*entity.Guest, int64, error)
	Delete(ctx context.Context, id, invitationID string) error
}
