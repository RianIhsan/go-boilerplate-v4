package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	authdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/entity"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/auth/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUsecase_Register(t *testing.T) {
	type fields struct {
		setupMock func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService)
	}
	type args struct {
		ctx context.Context
		req *authdto.RegisterRequest
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - new user registered",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "john@example.com").
						Return(nil, errors.New("not found"))
					userRepo.EXPECT().
						Create(gomock.Any(), gomock.Any()).
						Return(nil)
					jwtSvc.EXPECT().
						GenerateToken(gomock.Any(), "john@example.com").
						Return("mock.jwt.token", nil)
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.RegisterRequest{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
				},
			},
			wantErr: false,
		},
		{
			name: "error - email already exists",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "existing@example.com").
						Return(&entity.User{
							ID:    "existing-id",
							Email: "existing@example.com",
						}, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.RegisterRequest{
					Name:     "Existing",
					Email:    "existing@example.com",
					Password: "password123",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.UserConflict("existing@example.com"),
		},
		{
			name: "error - repository create fails",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "new@example.com").
						Return(nil, errors.New("not found"))
					userRepo.EXPECT().
						Create(gomock.Any(), gomock.Any()).
						Return(errors.New("db error"))
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.RegisterRequest{
					Name:     "New User",
					Email:    "new@example.com",
					Password: "password123",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
		{
			name: "error - email conflict from racing duplicate insert",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					// FindByEmail pre-check passes (no existing row yet), but
					// a concurrent registration with the same email wins the
					// race and the DB's unique constraint rejects this insert.
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "racer@example.com").
						Return(nil, errors.New("not found"))
					userRepo.EXPECT().
						Create(gomock.Any(), gomock.Any()).
						Return(apperrors.ErrConflict)
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.RegisterRequest{
					Name:     "Racer",
					Email:    "racer@example.com",
					Password: "password123",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.UserConflict("racer@example.com"),
		},
		{
			name: "error - hash password fails",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "toolong@example.com").
						Return(nil, errors.New("not found"))
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.RegisterRequest{
					Name:  "Too Long",
					Email: "toolong@example.com",
					// bcrypt rejects passwords longer than 72 bytes.
					Password: strings.Repeat("a", 73),
				},
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
		{
			name: "error - token generation fails",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "tokenfail@example.com").
						Return(nil, errors.New("not found"))
					userRepo.EXPECT().
						Create(gomock.Any(), gomock.Any()).
						Return(nil)
					jwtSvc.EXPECT().
						GenerateToken(gomock.Any(), "tokenfail@example.com").
						Return("", errors.New("signing error"))
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.RegisterRequest{
					Name:     "Token Fail",
					Email:    "tokenfail@example.com",
					Password: "password123",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock.NewMockUserRepository(ctrl)
			mockJWTSvc := mock.NewMockJWTService(ctrl)
			tt.fields.setupMock(mockUserRepo, mockJWTSvc)

			uc := usecase.NewAuthUsecase(mockUserRepo, mockJWTSvc)
			got, err := uc.Register(tt.args.ctx, tt.args.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Register() expected error but got nil")
					return
				}
				if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("Register() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Register() unexpected error = %v", err)
				return
			}
			if got == nil {
				t.Error("Register() returned nil response")
				return
			}
			if got.AccessToken == "" {
				t.Error("Register() returned empty access token")
			}
		})
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate test password hash: %v", err)
	}
	hashedPassword := string(hashed)

	type fields struct {
		setupMock func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService)
	}
	type args struct {
		ctx context.Context
		req *authdto.LoginRequest
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - valid credentials",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "john@example.com").
						Return(&entity.User{
							ID:        "user-id-1",
							Email:     "john@example.com",
							Password:  hashedPassword,
							CreatedAt: time.Now(),
						}, nil)
					jwtSvc.EXPECT().
						GenerateToken("user-id-1", "john@example.com").
						Return("mock.jwt.token", nil)
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.LoginRequest{
					Email:    "john@example.com",
					Password: "password123",
				},
			},
			wantErr: false,
		},
		{
			name: "error - user not found",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "notfound@example.com").
						Return(nil, apperrors.ErrNotFound)
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.LoginRequest{
					Email:    "notfound@example.com",
					Password: "password123",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInvalidCredential,
		},
		{
			name: "error - wrong password",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "john@example.com").
						Return(&entity.User{
							ID:       "user-id-1",
							Email:    "john@example.com",
							Password: hashedPassword,
						}, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.LoginRequest{
					Email:    "john@example.com",
					Password: "wrongpassword",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInvalidCredential,
		},
		{
			name: "error - token generation fails",
			fields: fields{
				setupMock: func(userRepo *mock.MockUserRepository, jwtSvc *mock.MockJWTService) {
					userRepo.EXPECT().
						FindByEmail(gomock.Any(), "john@example.com").
						Return(&entity.User{
							ID:       "user-id-1",
							Email:    "john@example.com",
							Password: hashedPassword,
						}, nil)
					jwtSvc.EXPECT().
						GenerateToken("user-id-1", "john@example.com").
						Return("", errors.New("signing error"))
				},
			},
			args: args{
				ctx: context.Background(),
				req: &authdto.LoginRequest{
					Email:    "john@example.com",
					Password: "password123",
				},
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock.NewMockUserRepository(ctrl)
			mockJWTSvc := mock.NewMockJWTService(ctrl)
			tt.fields.setupMock(mockUserRepo, mockJWTSvc)

			uc := usecase.NewAuthUsecase(mockUserRepo, mockJWTSvc)
			got, err := uc.Login(tt.args.ctx, tt.args.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() expected error but got nil")
					return
				}
				if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("Login() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Login() unexpected error = %v", err)
				return
			}
			if got == nil || got.AccessToken == "" {
				t.Error("Login() returned empty response")
			}
		})
	}
}
