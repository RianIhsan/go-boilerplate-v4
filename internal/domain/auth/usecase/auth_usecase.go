package usecase

import (
	"context"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/dto"
)

//go:generate mockgen -source=auth_usecase.go -destination=../../../mock/mock_auth_usecase.go -package=mock

type AuthUsecase interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
}
