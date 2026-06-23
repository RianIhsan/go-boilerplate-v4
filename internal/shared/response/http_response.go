package response

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var log *zap.Logger

// SetLogger lets main wire a logger so 5xx errors get the real cause logged
// server-side, correlated by the same trace_id returned to the client.
func SetLogger(l *zap.Logger) { log = l }

// Meta carries pagination info for list responses.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewMeta builds Meta from already-computed pagination values.
func NewMeta(page, limit int, total int64, totalPages int) *Meta {
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ErrorField is a single field-level validation failure.
type ErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value"`
}

// ErrorBody is the `error` object included in every error response.
type ErrorBody struct {
	Code    string       `json:"code"`
	Details string       `json:"details"`
	Fields  []ErrorField `json:"fields,omitempty"`
	TraceID string       `json:"trace_id"`
}

type successBody struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    *Meta  `json:"meta"`
}

type errorResponseBody struct {
	Success bool      `json:"success"`
	Status  int       `json:"status"`
	Message string    `json:"message"`
	Error   ErrorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Success sends a single-resource success response (GET by ID, POST, PUT).
func Success(w http.ResponseWriter, status int, message string, data any) {
	writeJSON(w, status, successBody{
		Success: true,
		Status:  status,
		Message: message,
		Data:    data,
		Meta:    nil,
	})
}

// SuccessList sends a success response for a paginated list (GET index).
func SuccessList(w http.ResponseWriter, message string, data any, meta *Meta) {
	writeJSON(w, http.StatusOK, successBody{
		Success: true,
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// NoContent sends a 204 with no body, the preferred shape for DELETE.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error maps an error to the standard error envelope. Non-AppError values are
// treated as internal errors. 5xx causes are logged server-side with the
// request's trace_id; only safe Details ever reach the client.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		appErr = apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	traceID := chimiddleware.GetReqID(r.Context())

	if appErr.Code >= http.StatusInternalServerError && log != nil {
		log.Error("internal error",
			zap.String("trace_id", traceID),
			zap.String("code", appErr.ErrCode),
			zap.Error(appErr.Cause),
		)
	}

	writeJSON(w, appErr.Code, errorResponseBody{
		Success: false,
		Status:  appErr.Code,
		Message: appErr.Message,
		Error: ErrorBody{
			Code:    appErr.ErrCode,
			Details: appErr.Details,
			TraceID: traceID,
		},
	})
}

// ValidationError sends a 422 with per-field validation failures.
func ValidationError(w http.ResponseWriter, r *http.Request, fields []ErrorField) {
	writeJSON(w, http.StatusUnprocessableEntity, errorResponseBody{
		Success: false,
		Status:  http.StatusUnprocessableEntity,
		Message: "validation failed",
		Error: ErrorBody{
			Code:    "VALIDATION_ERROR",
			Details: "request body contains invalid fields",
			Fields:  fields,
			TraceID: chimiddleware.GetReqID(r.Context()),
		},
	})
}
