package usecase

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/dto"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=invitation_usecase.go -destination=../../../mock/mock_invitation_usecase.go -package=mock

type InvitationUsecase interface {
	Create(ctx context.Context, userID string, req *dto.CreateInvitationRequest) (*dto.InvitationResponse, error)
	GetByID(ctx context.Context, id, userID string) (*dto.InvitationResponse, error)
	GetAll(ctx context.Context, userID string, pg pagination.Pagination) (*dto.InvitationListResponse, error)
	Update(ctx context.Context, id, userID string, req *dto.UpdateInvitationRequest) (*dto.InvitationResponse, error)
	Delete(ctx context.Context, id, userID string) error
	GetPublicBySlug(ctx context.Context, slug string) (*dto.PublicInvitationResponse, error)
}
