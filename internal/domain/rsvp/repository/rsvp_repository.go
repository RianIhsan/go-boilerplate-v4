package repository

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/entity"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=rsvp_repository.go -destination=../../../mock/mock_rsvp_repository.go -package=mock

type RSVPRepository interface {
	Create(ctx context.Context, rsvp *entity.RSVP) error
	Update(ctx context.Context, rsvp *entity.RSVP) error
	FindByGuestID(ctx context.Context, guestID string) (*entity.RSVP, error)
	FindAllByInvitationID(ctx context.Context, invitationID string, pg pagination.Pagination) ([]*entity.RSVP, int64, error)
}
