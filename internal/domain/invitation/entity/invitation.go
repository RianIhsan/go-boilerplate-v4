package entity

import "time"

type InvitationStatus string

const (
	InvitationStatusDraft   InvitationStatus = "draft"
	InvitationStatusActive  InvitationStatus = "active"
	InvitationStatusExpired InvitationStatus = "expired"
)

type Invitation struct {
	ID           string
	UserID       string
	Title        string
	Slug         string
	EventType    string
	EventDate    time.Time
	VenueName    string
	VenueAddress string
	VenueLat     *float64
	VenueLng     *float64
	Status       InvitationStatus
	IsPublished  bool
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
