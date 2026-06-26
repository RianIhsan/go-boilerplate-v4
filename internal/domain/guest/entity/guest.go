package entity

import "time"

type Guest struct {
	ID           string
	InvitationID string
	Name         string
	Phone        string
	Email        string
	UniqueToken  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
