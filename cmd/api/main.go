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
	filemetahandler "github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/handler"
	filemetausecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/usecase"
	todohandler "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/handler"
	todousecase "github.com/RianIhsan/go-boilerplate-v4/internal/domain/todo/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/infrastructure/database"
	"github.com/RianIhsan/go-boilerplate-v4/internal/infrastructure/persistence"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/middleware"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/jwt"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	log, err := logger.NewLogger(cfg.App.Env)
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}
	defer func() { _ = log.Sync() }()
	response.SetLogger(log)

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

	jwtSvc := jwt.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.ExpirationHours)

	userRepo := persistence.NewUserRepository(db)
	todoRepo := persistence.NewTodoRepository(db)

	authUC := authusecase.NewAuthUsecase(userRepo, jwtSvc)
	todoUC := todousecase.NewTodoUsecase(todoRepo)
	filemetaUC := filemetausecase.NewFileMetaUsecase()

	authH := authhandler.NewAuthHandler(authUC)
	todoH := todohandler.NewTodoHandler(todoUC)
	filemetaH := filemetahandler.NewFileMetaHandler(filemetaUC)

	if err := os.MkdirAll(constants.UploadTempDir, 0750); err != nil {
		log.Fatal(fmt.Sprintf("failed to create upload temp dir: %v", err))
	}

	mw := middleware.NewManager(log, jwtSvc)

	r := chi.NewRouter()
	mw.Apply(r)

	r.Route("/api/v1", func(r chi.Router) {
		authhandler.RegisterRoutes(r, authH, mw.AuthRateLimit())
		todohandler.RegisterRoutes(r, todoH, mw.Auth())
		filemetahandler.RegisterRoutes(r, filemetaH, mw.UploadRateLimit())
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
		if err := json.NewEncoder(w).Encode(map[string]string{"status": status}); err != nil {
			log.Error(fmt.Sprintf("failed to write health response: %v", err))
		}
	})

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
