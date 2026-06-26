package entity

import "time"

type RSVPStatus string

const (
	RSVPStatusAttending    RSVPStatus = "attending"
	RSVPStatusNotAttending RSVPStatus = "not_attending"
	RSVPStatusMaybe        RSVPStatus = "maybe"
)

type RSVP struct {
	ID            string
	InvitationID  string
	GuestID       *string
	Name          string
	Status        RSVPStatus
	AttendeeCount int
	Message       string
	RespondedAt   time.Time
}
