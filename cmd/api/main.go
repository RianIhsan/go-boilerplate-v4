package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/RianIhsan/go-boilerplate-v4/config"
	authhandler "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/handler"
	authusecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/usecase"
	todohandler "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/handler"
	todousecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/infrastructure/database"
	"github.com/RianIhsan/go-boilerplate-v4/internal/infrastructure/persistence"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/middleware"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/jwt"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/logger"
)

func main() {
	// ── Config ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// ── Logger ──────────────────────────────────────────────
	log, err := logger.NewLogger(cfg.App.Env)
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}
	defer log.Sync()
	response.SetLogger(log)

	// ── Database ─────────────────────────────────────────────
	db, err := database.NewPostgres(database.PostgresConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to connect to database: %v", err))
	}
	defer db.Close()

	// ── Services ─────────────────────────────────────────────
	jwtSvc := jwt.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.ExpirationHours)

	// ── Repositories ─────────────────────────────────────────
	userRepo := persistence.NewUserRepository(db)
	todoRepo := persistence.NewTodoRepository(db)

	// ── Usecases ─────────────────────────────────────────────
	authUC := authusecase.NewAuthUsecase(userRepo, jwtSvc)
	todoUC := todousecase.NewTodoUsecase(todoRepo)

	// ── Handlers ─────────────────────────────────────────────
	authH := authhandler.NewAuthHandler(authUC)
	todoH := todohandler.NewTodoHandler(todoUC)

	// ── Middleware ───────────────────────────────────────────
	mw := middleware.NewManager(log, jwtSvc)

	// ── Router ───────────────────────────────────────────────
	r := chi.NewRouter()
	mw.Apply(r)

	r.Route("/api/v1", func(r chi.Router) {
		authhandler.RegisterRoutes(r, authH)
		todohandler.RegisterRoutes(r, todoH, mw.Auth())
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status, code := "ok", http.StatusOK
		if err := db.PingContext(ctx); err != nil {
			status, code = "db unavailable", http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"status": status})
	})

	// ── Server ───────────────────────────────────────────────
	server := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	go func() {
		log.Info(fmt.Sprintf("server running on port %s", cfg.App.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(fmt.Sprintf("server failed: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("forced shutdown: %v", err))
	}

	log.Info("server exited gracefully")
}
