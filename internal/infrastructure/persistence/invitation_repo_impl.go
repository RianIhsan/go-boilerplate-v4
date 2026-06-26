package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/entity"
	invrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/lib/pq"
)

type invitationRepositoryImpl struct {
	db *sql.DB
}

func NewInvitationRepository(db *sql.DB) invrepo.InvitationRepository {
	return &invitationRepositoryImpl{db: db}
}

func (r *invitationRepositoryImpl) Create(ctx context.Context, invitation *entity.Invitation) error {
	query := `
		INSERT INTO invitations (
			id, user_id, title, slug, event_type, event_date, venue_name, venue_address,
			venue_lat, venue_lng, status, is_published, published_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		invitation.ID, invitation.UserID, invitation.Title, invitation.Slug, invitation.EventType,
		invitation.EventDate, invitation.VenueName, invitation.VenueAddress, invitation.VenueLat,
		invitation.VenueLng, invitation.Status, invitation.IsPublished, invitation.PublishedAt,
		invitation.CreatedAt, invitation.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pqUniqueViolation {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("invitationRepository.Create: %w", err)
	}
	return nil
}

func (r *invitationRepositoryImpl) FindByID(ctx context.Context, id, userID string) (*entity.Invitation, error) {
	query := `
		SELECT id, user_id, title, slug, event_type, event_date, venue_name, venue_address,
		       venue_lat, venue_lng, status, is_published, published_at, created_at, updated_at
		FROM invitations
		WHERE id = $1 AND user_id = $2
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, id, userID), "FindByID")
}

func (r *invitationRepositoryImpl) FindBySlug(ctx context.Context, slug string) (*entity.Invitation, error) {
	query := `
		SELECT id, user_id, title, slug, event_type, event_date, venue_name, venue_address,
		       venue_lat, venue_lng, status, is_published, published_at, created_at, updated_at
		FROM invitations
		WHERE slug = $1
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, slug), "FindBySlug")
}

func (r *invitationRepositoryImpl) scanRow(row *sql.Row, op string) (*entity.Invitation, error) {
	invitation := &entity.Invitation{}
	err := row.Scan(
		&invitation.ID, &invitation.UserID, &invitation.Title, &invitation.Slug, &invitation.EventType,
		&invitation.EventDate, &invitation.VenueName, &invitation.VenueAddress, &invitation.VenueLat,
		&invitation.VenueLng, &invitation.Status, &invitation.IsPublished, &invitation.PublishedAt,
		&invitation.CreatedAt, &invitation.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.%s: %w", op, err)
	}
	return invitation, nil
}

func (r *invitationRepositoryImpl) FindAllByUserID(ctx context.Context, userID string, pg pagination.Pagination) ([]*entity.Invitation, int64, error) {
	countQuery := `SELECT COUNT(*) FROM invitations WHERE user_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("invitationRepository.FindAllByUserID count: %w", err)
	}

	query := `
		SELECT id, user_id, title, slug, event_type, event_date, venue_name, venue_address,
		       venue_lat, venue_lng, status, is_published, published_at, created_at, updated_at
		FROM invitations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pg.Limit, pg.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("invitationRepository.FindAllByUserID: %w", err)
	}
	defer rows.Close()

	var invitations []*entity.Invitation
	for rows.Next() {
		invitation := &entity.Invitation{}
		if err := rows.Scan(
			&invitation.ID, &invitation.UserID, &invitation.Title, &invitation.Slug, &invitation.EventType,
			&invitation.EventDate, &invitation.VenueName, &invitation.VenueAddress, &invitation.VenueLat,
			&invitation.VenueLng, &invitation.Status, &invitation.IsPublished, &invitation.PublishedAt,
			&invitation.CreatedAt, &invitation.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("invitationRepository.FindAllByUserID scan: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	return invitations, total, nil
}

func (r *invitationRepositoryImpl) Update(ctx context.Context, invitation *entity.Invitation) error {
	query := `
		UPDATE invitations
		SET title = $1, event_type = $2, event_date = $3, venue_name = $4, venue_address = $5,
		    venue_lat = $6, venue_lng = $7, status = $8, is_published = $9, published_at = $10, updated_at = $11
		WHERE id = $12 AND user_id = $13
	`
	result, err := r.db.ExecContext(ctx, query,
		invitation.Title, invitation.EventType, invitation.EventDate, invitation.VenueName, invitation.VenueAddress,
		invitation.VenueLat, invitation.VenueLng, invitation.Status, invitation.IsPublished, invitation.PublishedAt,
		invitation.UpdatedAt, invitation.ID, invitation.UserID,
	)
	if err != nil {
		return fmt.Errorf("invitationRepository.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *invitationRepositoryImpl) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM invitations WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("invitationRepository.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
