package dto

import "time"

type RSVPResponse struct {
	ID            string    `json:"id"`
	InvitationID  string    `json:"invitation_id"`
	GuestID       *string   `json:"guest_id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	AttendeeCount int       `json:"attendee_count"`
	Message       string    `json:"message"`
	RespondedAt   time.Time `json:"responded_at"`
}

type RSVPListResponse struct {
	Items      []*RSVPResponse `json:"items"`
	TotalItems int64           `json:"total_items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
