package errors

import "net/http"

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

func TodoNotFound(id string) *AppError {
	return New(http.StatusNotFound, "TODO_NOT_FOUND", "todo not found", "no todo with id "+id+" exists")
}

func UserConflict(email string) *AppError {
	return New(http.StatusConflict, "DUPLICATE_ENTRY", "user already exists", "a user with email "+email+" already exists")
}

func UnsupportedFileType() *AppError {
	return New(http.StatusUnprocessableEntity, "UNSUPPORTED_FILE_TYPE", "unsupported file type", "could not recognize file as one of: apk, exe, deb, rpm, dmg")
}

func FileTooLarge() *AppError {
	return New(http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file exceeds the maximum allowed size", "")
}

func FileParseFailed() *AppError {
	return New(http.StatusUnprocessableEntity, "FILE_PARSE_FAILED", "could not parse file metadata", "")
}
