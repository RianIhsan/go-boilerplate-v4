package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/entity"
	rsvprepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
)

type rsvpRepositoryImpl struct {
	db *sql.DB
}

func NewRSVPRepository(db *sql.DB) rsvprepo.RSVPRepository {
	return &rsvpRepositoryImpl{db: db}
}

func (r *rsvpRepositoryImpl) Create(ctx context.Context, rsvp *entity.RSVP) error {
	query := `
		INSERT INTO rsvps (id, invitation_id, guest_id, name, status, attendee_count, message, responded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		rsvp.ID, rsvp.InvitationID, rsvp.GuestID, rsvp.Name, rsvp.Status, rsvp.AttendeeCount,
		rsvp.Message, rsvp.RespondedAt,
	)
	if err != nil {
		return fmt.Errorf("rsvpRepository.Create: %w", err)
	}
	return nil
}

func (r *rsvpRepositoryImpl) Update(ctx context.Context, rsvp *entity.RSVP) error {
	query := `
		UPDATE rsvps
		SET status = $1, attendee_count = $2, message = $3, responded_at = $4
		WHERE id = $5
	`
	result, err := r.db.ExecContext(ctx, query,
		rsvp.Status, rsvp.AttendeeCount, rsvp.Message, rsvp.RespondedAt, rsvp.ID,
	)
	if err != nil {
		return fmt.Errorf("rsvpRepository.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *rsvpRepositoryImpl) FindByGuestID(ctx context.Context, guestID string) (*entity.RSVP, error) {
	query := `
		SELECT id, invitation_id, guest_id, name, status, attendee_count, message, responded_at
		FROM rsvps
		WHERE guest_id = $1
	`
	rsvp := &entity.RSVP{}
	err := r.db.QueryRowContext(ctx, query, guestID).Scan(
		&rsvp.ID, &rsvp.InvitationID, &rsvp.GuestID, &rsvp.Name, &rsvp.Status, &rsvp.AttendeeCount,
		&rsvp.Message, &rsvp.RespondedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("rsvpRepository.FindByGuestID: %w", err)
	}
	return rsvp, nil
}

func (r *rsvpRepositoryImpl) FindAllByInvitationID(ctx context.Context, invitationID string, pg pagination.Pagination) ([]*entity.RSVP, int64, error) {
	countQuery := `SELECT COUNT(*) FROM rsvps WHERE invitation_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, invitationID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("rsvpRepository.FindAllByInvitationID count: %w", err)
	}

	query := `
		SELECT id, invitation_id, guest_id, name, status, attendee_count, message, responded_at
		FROM rsvps
		WHERE invitation_id = $1
		ORDER BY responded_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, invitationID, pg.Limit, pg.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("rsvpRepository.FindAllByInvitationID: %w", err)
	}
	defer rows.Close()

	var rsvps []*entity.RSVP
	for rows.Next() {
		rsvp := &entity.RSVP{}
		if err := rows.Scan(
			&rsvp.ID, &rsvp.InvitationID, &rsvp.GuestID, &rsvp.Name, &rsvp.Status, &rsvp.AttendeeCount,
			&rsvp.Message, &rsvp.RespondedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("rsvpRepository.FindAllByInvitationID scan: %w", err)
		}
		rsvps = append(rsvps, rsvp)
	}

	return rsvps, total, nil
}
