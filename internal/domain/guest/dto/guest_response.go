package dto

import "time"

type GuestResponse struct {
	ID           string    `json:"id"`
	InvitationID string    `json:"invitation_id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	UniqueToken  string    `json:"unique_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GuestListResponse struct {
	Items      []*GuestResponse `json:"items"`
	TotalItems int64            `json:"total_items"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type PublicGuestResponse struct {
	GuestName       string    `json:"guest_name"`
	InvitationTitle string    `json:"invitation_title"`
	EventType       string    `json:"event_type"`
	EventDate       time.Time `json:"event_date"`
	VenueName       string    `json:"venue_name"`
	VenueAddress    string    `json:"venue_address"`
	VenueLat        *float64  `json:"venue_lat"`
	VenueLng        *float64  `json:"venue_lng"`
}
