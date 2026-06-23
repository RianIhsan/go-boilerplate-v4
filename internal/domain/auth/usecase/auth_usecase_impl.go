package usecase

import (
	"context"
	"time"

	authdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/entity"
	authrepo "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/repository"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/crypto"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/jwt"
	"github.com/google/uuid"
)

type authUsecaseImpl struct {
	userRepo authrepo.UserRepository
	jwtSvc   jwt.JWTService
}

func NewAuthUsecase(userRepo authrepo.UserRepository, jwtSvc jwt.JWTService) AuthUsecase {
	return &authUsecaseImpl{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
	}
}

func (u *authUsecaseImpl) Register(ctx context.Context, req *authdto.RegisterRequest) (*authdto.AuthResponse, error) {
	existing, _ := u.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, apperrors.UserConflict(req.Email)
	}

	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	user := &entity.User{
		ID:        uuid.NewString(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	token, err := u.jwtSvc.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return &authdto.AuthResponse{
		AccessToken: token,
		User: authdto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (u *authUsecaseImpl) Login(ctx context.Context, req *authdto.LoginRequest) (*authdto.AuthResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, apperrors.ErrInvalidCredential
	}

	if !crypto.CheckPasswordHash(req.Password, user.Password) {
		return nil, apperrors.ErrInvalidCredential
	}

	token, err := u.jwtSvc.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	return &authdto.AuthResponse{
		AccessToken: token,
		User: authdto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
