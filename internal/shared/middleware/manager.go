package middleware

import (
	"net/http"

	"github.com/RianIhsan/go-boilerplate-v4/pkg/jwt"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Manager is the single place new middlewares get registered. Global,
// always-on middlewares go in Apply; middlewares that only apply to specific
// route groups (like Auth) get their own accessor so handlers' RegisterRoutes
// can pull them in explicitly.
type Manager interface {
	Apply(r chi.Router)
	Auth() func(http.Handler) http.Handler
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
