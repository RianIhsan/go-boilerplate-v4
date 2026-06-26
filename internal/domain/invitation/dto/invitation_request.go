package dto

import "time"

type CreateInvitationRequest struct {
	Title        string    `json:"title"         validate:"required,min=1,max=255"`
	Slug         string    `json:"slug"          validate:"omitempty,min=3,max=100"`
	EventType    string    `json:"event_type"    validate:"required,min=1,max=50"`
	EventDate    time.Time `json:"event_date"    validate:"required"`
	VenueName    string    `json:"venue_name"    validate:"required,min=1,max=255"`
	VenueAddress string    `json:"venue_address" validate:"required,max=2000"`
	VenueLat     *float64  `json:"venue_lat"     validate:"omitempty,min=-90,max=90"`
	VenueLng     *float64  `json:"venue_lng"     validate:"omitempty,min=-180,max=180"`
}

type UpdateInvitationRequest struct {
	Title        string    `json:"title"         validate:"omitempty,min=1,max=255"`
	EventType    string    `json:"event_type"    validate:"omitempty,min=1,max=50"`
	EventDate    time.Time `json:"event_date"    validate:"omitempty"`
	VenueName    string    `json:"venue_name"    validate:"omitempty,min=1,max=255"`
	VenueAddress string    `json:"venue_address" validate:"omitempty,max=2000"`
	VenueLat     *float64  `json:"venue_lat"     validate:"omitempty,min=-90,max=90"`
	VenueLng     *float64  `json:"venue_lng"     validate:"omitempty,min=-180,max=180"`
	IsPublished  *bool     `json:"is_published"  validate:"omitempty"`
}
