package validator

import (
	"fmt"
	"strings"

	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Validate(s interface{}) []response.ErrorField {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var fields []response.ErrorField
	for _, e := range err.(validator.ValidationErrors) {
		fields = append(fields, response.ErrorField{
			Field:   strings.ToLower(e.Field()),
			Message: formatError(e),
			Value:   safeValue(e),
		})
	}
	return fields
}

func safeValue(e validator.FieldError) any {
	if strings.Contains(strings.ToLower(e.Field()), "password") {
		return nil
	}
	return e.Value()
}

func formatError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
