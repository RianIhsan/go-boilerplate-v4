package errors

import "net/http"

// AppError is the internal error type returned by usecases. Details is safe to
// expose to API clients; Cause (set via Wrap) is the real underlying error and
// is only ever written to server-side logs, never serialized.
type AppError struct {
	Code    int    `json:"-"`
	ErrCode string `json:"-"`
	Message string `json:"-"`
	Details string `json:"-"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.Cause }

func New(code int, errCode, message, details string) *AppError {
	return &AppError{Code: code, ErrCode: errCode, Message: message, Details: details}
}

// Wrap attaches the real underlying cause to a base AppError for server-side
// logging, without changing what gets sent to the client.
func Wrap(base *AppError, cause error) *AppError {
	wrapped := *base
	wrapped.Cause = cause
	return &wrapped
}

var (
	ErrNotFound          = New(http.StatusNotFound, "NOT_FOUND", "resource not found", "")
	ErrUnauthorized      = New(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", "")
	ErrForbidden         = New(http.StatusForbidden, "FORBIDDEN", "forbidden", "")
	ErrBadRequest        = New(http.StatusBadRequest, "BAD_REQUEST", "invalid request body", "")
	ErrConflict          = New(http.StatusConflict, "DUPLICATE_ENTRY", "resource already exists", "")
	ErrInternalServer    = New(http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred. please try again later", "")
	ErrInvalidToken      = New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token", "")
	ErrInvalidCredential = New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", "")
	ErrTooManyRequests   = New(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "too many requests, please try again later", "")
)

// TodoNotFound builds a 404 with details scoped to the specific todo id.
func TodoNotFound(id string) *AppError {
	return New(http.StatusNotFound, "TODO_NOT_FOUND", "todo not found", "no todo with id "+id+" exists")
}

// UserConflict builds a 409 with details scoped to the specific email.
func UserConflict(email string) *AppError {
	return New(http.StatusConflict, "DUPLICATE_ENTRY", "user already exists", "a user with email "+email+" already exists")
}
