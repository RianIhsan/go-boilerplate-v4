package usecase

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/dto"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=guest_usecase.go -destination=../../../mock/mock_guest_usecase.go -package=mock

type GuestUsecase interface {
	Create(ctx context.Context, userID, invitationID string, req *dto.CreateGuestRequest) (*dto.GuestResponse, error)
	GetAll(ctx context.Context, userID, invitationID string, pg pagination.Pagination) (*dto.GuestListResponse, error)
	Delete(ctx context.Context, userID, invitationID, guestID string) error
	GetPublicByToken(ctx context.Context, slug, token string) (*dto.PublicGuestResponse, error)
}
