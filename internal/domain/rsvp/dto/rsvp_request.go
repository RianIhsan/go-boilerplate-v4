package dto

type SubmitRSVPRequest struct {
	GuestToken    string `json:"guest_token"    validate:"omitempty,len=64"`
	Name          string `json:"name"           validate:"required_without=GuestToken,max=150"`
	Status        string `json:"status"         validate:"required,oneof=attending not_attending maybe"`
	AttendeeCount int    `json:"attendee_count" validate:"omitempty,min=1,max=20"`
	Message       string `json:"message"        validate:"omitempty,max=1000"`
}
