package dto

type CreateTodoRequest struct {
	Title       string `json:"title"       validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateTodoRequest struct {
	Title       string `json:"title"       validate:"omitempty,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Status      string `json:"status"      validate:"omitempty,oneof=pending in_progress done"`
}
