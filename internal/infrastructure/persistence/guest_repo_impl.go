package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/entity"
	guestrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/lib/pq"
)

type guestRepositoryImpl struct {
	db *sql.DB
}

func NewGuestRepository(db *sql.DB) guestrepo.GuestRepository {
	return &guestRepositoryImpl{db: db}
}

func (r *guestRepositoryImpl) Create(ctx context.Context, guest *entity.Guest) error {
	query := `
		INSERT INTO guests (id, invitation_id, name, phone, email, unique_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		guest.ID, guest.InvitationID, guest.Name, guest.Phone, guest.Email, guest.UniqueToken,
		guest.CreatedAt, guest.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pqUniqueViolation {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("guestRepository.Create: %w", err)
	}
	return nil
}

func (r *guestRepositoryImpl) FindByID(ctx context.Context, id, invitationID string) (*entity.Guest, error) {
	query := `
		SELECT id, invitation_id, name, phone, email, unique_token, created_at, updated_at
		FROM guests
		WHERE id = $1 AND invitation_id = $2
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, id, invitationID), "FindByID")
}

func (r *guestRepositoryImpl) FindByToken(ctx context.Context, token string) (*entity.Guest, error) {
	query := `
		SELECT id, invitation_id, name, phone, email, unique_token, created_at, updated_at
		FROM guests
		WHERE unique_token = $1
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, token), "FindByToken")
}

func (r *guestRepositoryImpl) scanRow(row *sql.Row, op string) (*entity.Guest, error) {
	guest := &entity.Guest{}
	err := row.Scan(
		&guest.ID, &guest.InvitationID, &guest.Name, &guest.Phone, &guest.Email, &guest.UniqueToken,
		&guest.CreatedAt, &guest.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("guestRepository.%s: %w", op, err)
	}
	return guest, nil
}

func (r *guestRepositoryImpl) FindAllByInvitationID(ctx context.Context, invitationID string, pg pagination.Pagination) ([]*entity.Guest, int64, error) {
	countQuery := `SELECT COUNT(*) FROM guests WHERE invitation_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, invitationID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("guestRepository.FindAllByInvitationID count: %w", err)
	}

	query := `
		SELECT id, invitation_id, name, phone, email, unique_token, created_at, updated_at
		FROM guests
		WHERE invitation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, invitationID, pg.Limit, pg.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("guestRepository.FindAllByInvitationID: %w", err)
	}
	defer rows.Close()

	var guests []*entity.Guest
	for rows.Next() {
		guest := &entity.Guest{}
		if err := rows.Scan(
			&guest.ID, &guest.InvitationID, &guest.Name, &guest.Phone, &guest.Email, &guest.UniqueToken,
			&guest.CreatedAt, &guest.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("guestRepository.FindAllByInvitationID scan: %w", err)
		}
		guests = append(guests, guest)
	}

	return guests, total, nil
}

func (r *guestRepositoryImpl) Delete(ctx context.Context, id, invitationID string) error {
	query := `DELETE FROM guests WHERE id = $1 AND invitation_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, invitationID)
	if err != nil {
		return fmt.Errorf("guestRepository.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
