package usecase

import (
	"context"
	"time"

	guestdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/entity"
	guestrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/repository"
	invrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/random"
	"github.com/google/uuid"
)

type guestUsecaseImpl struct {
	guestRepo      guestrepo.GuestRepository
	invitationRepo invrepo.InvitationRepository
}

func NewGuestUsecase(guestRepo guestrepo.GuestRepository, invitationRepo invrepo.InvitationRepository) GuestUsecase {
	return &guestUsecaseImpl{guestRepo: guestRepo, invitationRepo: invitationRepo}
}

func (u *guestUsecaseImpl) Create(ctx context.Context, userID, invitationID string, req *guestdto.CreateGuestRequest) (*guestdto.GuestResponse, error) {
	if _, err := u.invitationRepo.FindByID(ctx, invitationID, userID); err != nil {
		return nil, apperrors.InvitationNotFound(invitationID)
	}

	token, err := random.Hex(32)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	guest := &entity.Guest{
		ID:           uuid.NewString(),
		InvitationID: invitationID,
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		UniqueToken:  token,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.guestRepo.Create(ctx, guest); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return toResponse(guest), nil
}

func (u *guestUsecaseImpl) GetAll(ctx context.Context, userID, invitationID string, pg pagination.Pagination) (*guestdto.GuestListResponse, error) {
	if _, err := u.invitationRepo.FindByID(ctx, invitationID, userID); err != nil {
		return nil, apperrors.InvitationNotFound(invitationID)
	}

	guests, total, err := u.guestRepo.FindAllByInvitationID(ctx, invitationID, pg)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	items := make([]*guestdto.GuestResponse, 0, len(guests))
	for _, g := range guests {
		items = append(items, toResponse(g))
	}

	return &guestdto.GuestListResponse{
		Items:      items,
		TotalItems: total,
		Page:       pg.Page,
		Limit:      pg.Limit,
		TotalPages: pagination.TotalPages(total, pg.Limit),
	}, nil
}

func (u *guestUsecaseImpl) Delete(ctx context.Context, userID, invitationID, guestID string) error {
	if _, err := u.invitationRepo.FindByID(ctx, invitationID, userID); err != nil {
		return apperrors.InvitationNotFound(invitationID)
	}

	if err := u.guestRepo.Delete(ctx, guestID, invitationID); err != nil {
		return apperrors.GuestNotFound(guestID)
	}
	return nil
}

func (u *guestUsecaseImpl) GetPublicByToken(ctx context.Context, slug, token string) (*guestdto.PublicGuestResponse, error) {
	invitation, err := u.invitationRepo.FindBySlug(ctx, slug)
	if err != nil || !invitation.IsPublished {
		return nil, apperrors.ErrNotFound
	}

	guest, err := u.guestRepo.FindByToken(ctx, token)
	if err != nil || guest.InvitationID != invitation.ID {
		return nil, apperrors.ErrNotFound
	}

	return &guestdto.PublicGuestResponse{
		GuestName:       guest.Name,
		InvitationTitle: invitation.Title,
		EventType:       invitation.EventType,
		EventDate:       invitation.EventDate,
		VenueName:       invitation.VenueName,
		VenueAddress:    invitation.VenueAddress,
		VenueLat:        invitation.VenueLat,
		VenueLng:        invitation.VenueLng,
	}, nil
}

func toResponse(g *entity.Guest) *guestdto.GuestResponse {
	return &guestdto.GuestResponse{
		ID:           g.ID,
		InvitationID: g.InvitationID,
		Name:         g.Name,
		Phone:        g.Phone,
		Email:        g.Email,
		UniqueToken:  g.UniqueToken,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}
