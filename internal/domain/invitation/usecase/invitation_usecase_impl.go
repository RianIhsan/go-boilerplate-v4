package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	invdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/entity"
	invrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/random"
	"github.com/google/uuid"
)

var slugSanitizeRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugSanitizeRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func computeStatus(invitation *entity.Invitation) entity.InvitationStatus {
	if !invitation.IsPublished {
		return entity.InvitationStatusDraft
	}
	if time.Now().After(invitation.EventDate) {
		return entity.InvitationStatusExpired
	}
	return entity.InvitationStatusActive
}

type invitationUsecaseImpl struct {
	invitationRepo invrepo.InvitationRepository
}

func NewInvitationUsecase(invitationRepo invrepo.InvitationRepository) InvitationUsecase {
	return &invitationUsecaseImpl{invitationRepo: invitationRepo}
}

func (u *invitationUsecaseImpl) Create(ctx context.Context, userID string, req *invdto.CreateInvitationRequest) (*invdto.InvitationResponse, error) {
	userSuppliedSlug := req.Slug != ""

	slug := slugify(req.Slug)
	if slug == "" {
		slug = slugify(req.Title)
	}
	if slug == "" {
		slug = "undangan"
	}

	invitation := &entity.Invitation{
		ID:           uuid.NewString(),
		UserID:       userID,
		Title:        req.Title,
		Slug:         slug,
		EventType:    req.EventType,
		EventDate:    req.EventDate,
		VenueName:    req.VenueName,
		VenueAddress: req.VenueAddress,
		VenueLat:     req.VenueLat,
		VenueLng:     req.VenueLng,
		Status:       entity.InvitationStatusDraft,
		IsPublished:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.invitationRepo.Create(ctx, invitation); err != nil {
		if !errors.Is(err, apperrors.ErrConflict) {
			return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
		}
		if userSuppliedSlug {
			return nil, apperrors.SlugConflict(slug)
		}

		suffix, rerr := random.Hex(3)
		if rerr != nil {
			return nil, apperrors.Wrap(apperrors.ErrInternalServer, rerr)
		}
		invitation.Slug = slug + "-" + suffix
		if err := u.invitationRepo.Create(ctx, invitation); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
		}
	}

	return toResponse(invitation), nil
}

func (u *invitationUsecaseImpl) GetByID(ctx context.Context, id, userID string) (*invdto.InvitationResponse, error) {
	invitation, err := u.invitationRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, apperrors.InvitationNotFound(id)
	}

	u.syncStatus(ctx, invitation)
	return toResponse(invitation), nil
}

func (u *invitationUsecaseImpl) GetAll(ctx context.Context, userID string, pg pagination.Pagination) (*invdto.InvitationListResponse, error) {
	invitations, total, err := u.invitationRepo.FindAllByUserID(ctx, userID, pg)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	items := make([]*invdto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		inv.Status = computeStatus(inv)
		items = append(items, toResponse(inv))
	}

	return &invdto.InvitationListResponse{
		Items:      items,
		TotalItems: total,
		Page:       pg.Page,
		Limit:      pg.Limit,
		TotalPages: pagination.TotalPages(total, pg.Limit),
	}, nil
}

func (u *invitationUsecaseImpl) Update(ctx context.Context, id, userID string, req *invdto.UpdateInvitationRequest) (*invdto.InvitationResponse, error) {
	invitation, err := u.invitationRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, apperrors.InvitationNotFound(id)
	}

	if req.Title != "" {
		invitation.Title = req.Title
	}
	if req.EventType != "" {
		invitation.EventType = req.EventType
	}
	if !req.EventDate.IsZero() {
		invitation.EventDate = req.EventDate
	}
	if req.VenueName != "" {
		invitation.VenueName = req.VenueName
	}
	if req.VenueAddress != "" {
		invitation.VenueAddress = req.VenueAddress
	}
	if req.VenueLat != nil {
		invitation.VenueLat = req.VenueLat
	}
	if req.VenueLng != nil {
		invitation.VenueLng = req.VenueLng
	}
	if req.IsPublished != nil {
		invitation.IsPublished = *req.IsPublished
		if *req.IsPublished && invitation.PublishedAt == nil {
			now := time.Now()
			invitation.PublishedAt = &now
		}
	}

	invitation.Status = computeStatus(invitation)
	invitation.UpdatedAt = time.Now()

	if err := u.invitationRepo.Update(ctx, invitation); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return toResponse(invitation), nil
}

func (u *invitationUsecaseImpl) Delete(ctx context.Context, id, userID string) error {
	_, err := u.invitationRepo.FindByID(ctx, id, userID)
	if err != nil {
		return apperrors.InvitationNotFound(id)
	}

	if err := u.invitationRepo.Delete(ctx, id, userID); err != nil {
		return apperrors.Wrap(apperrors.ErrInternalServer, err)
	}
	return nil
}

func (u *invitationUsecaseImpl) GetPublicBySlug(ctx context.Context, slug string) (*invdto.PublicInvitationResponse, error) {
	invitation, err := u.invitationRepo.FindBySlug(ctx, slug)
	if err != nil || !invitation.IsPublished {
		return nil, apperrors.ErrNotFound
	}

	return toPublicResponse(invitation), nil
}

func (u *invitationUsecaseImpl) syncStatus(ctx context.Context, invitation *entity.Invitation) {
	computed := computeStatus(invitation)
	if computed == invitation.Status {
		return
	}
	invitation.Status = computed
	invitation.UpdatedAt = time.Now()
	_ = u.invitationRepo.Update(ctx, invitation)
}

func toResponse(invitation *entity.Invitation) *invdto.InvitationResponse {
	return &invdto.InvitationResponse{
		ID:           invitation.ID,
		UserID:       invitation.UserID,
		Title:        invitation.Title,
		Slug:         invitation.Slug,
		EventType:    invitation.EventType,
		EventDate:    invitation.EventDate,
		VenueName:    invitation.VenueName,
		VenueAddress: invitation.VenueAddress,
		VenueLat:     invitation.VenueLat,
		VenueLng:     invitation.VenueLng,
		Status:       string(invitation.Status),
		IsPublished:  invitation.IsPublished,
		PublishedAt:  invitation.PublishedAt,
		CreatedAt:    invitation.CreatedAt,
		UpdatedAt:    invitation.UpdatedAt,
	}
}

func toPublicResponse(invitation *entity.Invitation) *invdto.PublicInvitationResponse {
	return &invdto.PublicInvitationResponse{
		Title:        invitation.Title,
		EventType:    invitation.EventType,
		EventDate:    invitation.EventDate,
		VenueName:    invitation.VenueName,
		VenueAddress: invitation.VenueAddress,
		VenueLat:     invitation.VenueLat,
		VenueLng:     invitation.VenueLng,
	}
}
