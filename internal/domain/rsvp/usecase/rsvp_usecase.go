package usecase

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/dto"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

//go:generate mockgen -source=rsvp_usecase.go -destination=../../../mock/mock_rsvp_usecase.go -package=mock

type RSVPUsecase interface {
	Submit(ctx context.Context, slug string, req *dto.SubmitRSVPRequest) (*dto.RSVPResponse, error)
	GetAll(ctx context.Context, userID, invitationID string, pg pagination.Pagination) (*dto.RSVPListResponse, error)
}
