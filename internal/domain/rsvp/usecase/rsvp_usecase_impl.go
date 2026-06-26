package usecase

import (
	"context"
	"time"

	guestrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/repository"
	invrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/repository"
	rsvpdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/entity"
	rsvprepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/google/uuid"
)

type rsvpUsecaseImpl struct {
	rsvpRepo       rsvprepo.RSVPRepository
	invitationRepo invrepo.InvitationRepository
	guestRepo      guestrepo.GuestRepository
}

func NewRSVPUsecase(rsvpRepo rsvprepo.RSVPRepository, invitationRepo invrepo.InvitationRepository, guestRepo guestrepo.GuestRepository) RSVPUsecase {
	return &rsvpUsecaseImpl{rsvpRepo: rsvpRepo, invitationRepo: invitationRepo, guestRepo: guestRepo}
}

func (u *rsvpUsecaseImpl) Submit(ctx context.Context, slug string, req *rsvpdto.SubmitRSVPRequest) (*rsvpdto.RSVPResponse, error) {
	invitation, err := u.invitationRepo.FindBySlug(ctx, slug)
	if err != nil || !invitation.IsPublished {
		return nil, apperrors.ErrNotFound
	}

	attendeeCount := req.AttendeeCount
	if attendeeCount == 0 {
		attendeeCount = 1
	}

	var guestID *string
	name := req.Name

	if req.GuestToken != "" {
		guest, err := u.guestRepo.FindByToken(ctx, req.GuestToken)
		if err != nil || guest.InvitationID != invitation.ID {
			return nil, apperrors.ErrNotFound
		}
		guestID = &guest.ID
		name = guest.Name

		existing, err := u.rsvpRepo.FindByGuestID(ctx, guest.ID)
		if err == nil && existing != nil {
			existing.Status = entity.RSVPStatus(req.Status)
			existing.AttendeeCount = attendeeCount
			existing.Message = req.Message
			existing.RespondedAt = time.Now()

			if err := u.rsvpRepo.Update(ctx, existing); err != nil {
				return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
			}
			return toResponse(existing), nil
		}
	}

	rsvp := &entity.RSVP{
		ID:            uuid.NewString(),
		InvitationID:  invitation.ID,
		GuestID:       guestID,
		Name:          name,
		Status:        entity.RSVPStatus(req.Status),
		AttendeeCount: attendeeCount,
		Message:       req.Message,
		RespondedAt:   time.Now(),
	}

	if err := u.rsvpRepo.Create(ctx, rsvp); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return toResponse(rsvp), nil
}

func (u *rsvpUsecaseImpl) GetAll(ctx context.Context, userID, invitationID string, pg pagination.Pagination) (*rsvpdto.RSVPListResponse, error) {
	if _, err := u.invitationRepo.FindByID(ctx, invitationID, userID); err != nil {
		return nil, apperrors.InvitationNotFound(invitationID)
	}

	rsvps, total, err := u.rsvpRepo.FindAllByInvitationID(ctx, invitationID, pg)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	items := make([]*rsvpdto.RSVPResponse, 0, len(rsvps))
	for _, rs := range rsvps {
		items = append(items, toResponse(rs))
	}

	return &rsvpdto.RSVPListResponse{
		Items:      items,
		TotalItems: total,
		Page:       pg.Page,
		Limit:      pg.Limit,
		TotalPages: pagination.TotalPages(total, pg.Limit),
	}, nil
}

func toResponse(rsvp *entity.RSVP) *rsvpdto.RSVPResponse {
	return &rsvpdto.RSVPResponse{
		ID:            rsvp.ID,
		InvitationID:  rsvp.InvitationID,
		GuestID:       rsvp.GuestID,
		Name:          rsvp.Name,
		Status:        string(rsvp.Status),
		AttendeeCount: rsvp.AttendeeCount,
		Message:       rsvp.Message,
		RespondedAt:   rsvp.RespondedAt,
	}
}
