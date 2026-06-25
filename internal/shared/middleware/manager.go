package middleware

import (
	"net/http"
	"time"

	"github.com/RianIhsan/go-boilerplate-v4/pkg/jwt"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const (
	AuthRateLimitRequests = 10
	AuthRateLimitWindow   = time.Minute
)

const (
	UploadRateLimitRequests = 5
	UploadRateLimitWindow   = time.Minute
)

type Manager interface {
	Apply(r chi.Router)
	Auth() func(http.Handler) http.Handler
	AuthRateLimit() func(http.Handler) http.Handler
	UploadRateLimit() func(http.Handler) http.Handler
}

type manager struct {
	log    *zap.Logger
	jwtSvc jwt.JWTService
}

func NewManager(log *zap.Logger, jwtSvc jwt.JWTService) Manager {
	return &manager{log: log, jwtSvc: jwtSvc}
}

func (m *manager) Apply(r chi.Router) {
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(Recovery(m.log))
	r.Use(Logger(m.log))
}

func (m *manager) Auth() func(http.Handler) http.Handler {
	return Auth(m.jwtSvc)
}

func (m *manager) AuthRateLimit() func(http.Handler) http.Handler {
	return RateLimit(AuthRateLimitRequests, AuthRateLimitWindow)
}

func (m *manager) UploadRateLimit() func(http.Handler) http.Handler {
	return RateLimit(UploadRateLimitRequests, UploadRateLimitWindow)
}
