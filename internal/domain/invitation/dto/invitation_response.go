package dto

import "time"

type InvitationResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	EventType    string     `json:"event_type"`
	EventDate    time.Time  `json:"event_date"`
	VenueName    string     `json:"venue_name"`
	VenueAddress string     `json:"venue_address"`
	VenueLat     *float64   `json:"venue_lat"`
	VenueLng     *float64   `json:"venue_lng"`
	Status       string     `json:"status"`
	IsPublished  bool       `json:"is_published"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type InvitationListResponse struct {
	Items      []*InvitationResponse `json:"items"`
	TotalItems int64                 `json:"total_items"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
}

type PublicInvitationResponse struct {
	Title        string    `json:"title"`
	EventType    string    `json:"event_type"`
	EventDate    time.Time `json:"event_date"`
	VenueName    string    `json:"venue_name"`
	VenueAddress string    `json:"venue_address"`
	VenueLat     *float64  `json:"venue_lat"`
	VenueLng     *float64  `json:"venue_lng"`
}
