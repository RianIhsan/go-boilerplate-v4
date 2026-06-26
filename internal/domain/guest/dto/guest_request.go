package dto

type CreateGuestRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=150"`
	Phone string `json:"phone" validate:"omitempty,max=20"`
	Email string `json:"email" validate:"omitempty,email,max=255"`
}
