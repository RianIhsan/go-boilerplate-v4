package usecase_test

import (
	"errors"
	"testing"
	"time"

	"context"

	invdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/entity"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/golang/mock/gomock"
)

var (
	mockUserID       = "user-id-1"
	mockOtherUserID  = "user-id-2"
	mockInvitationID = "invitation-id-1"
	mockInvitation   = &entity.Invitation{
		ID:           mockInvitationID,
		UserID:       mockUserID,
		Title:        "Pernikahan Budi & Ani",
		Slug:         "pernikahan-budi-ani",
		EventType:    "wedding",
		EventDate:    time.Now().Add(30 * 24 * time.Hour),
		VenueName:    "Gedung Serbaguna",
		VenueAddress: "Jl. Merdeka No. 1",
		Status:       entity.InvitationStatusDraft,
		IsPublished:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
)

func TestInvitationUsecase_Create(t *testing.T) {
	tests := []struct {
		name        string
		req         *invdto.CreateInvitationRequest
		setupMock   func(repo *mock.MockInvitationRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - auto-generated slug",
			req: &invdto.CreateInvitationRequest{
				Title:        "Pernikahan Budi & Ani",
				EventType:    "wedding",
				EventDate:    time.Now().Add(30 * 24 * time.Hour),
				VenueName:    "Gedung Serbaguna",
				VenueAddress: "Jl. Merdeka No. 1",
			},
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - user-supplied slug already taken",
			req: &invdto.CreateInvitationRequest{
				Title:        "Pernikahan Budi & Ani",
				Slug:         "sudah-dipakai",
				EventType:    "wedding",
				EventDate:    time.Now().Add(30 * 24 * time.Hour),
				VenueName:    "Gedung Serbaguna",
				VenueAddress: "Jl. Merdeka No. 1",
			},
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(apperrors.ErrConflict)
			},
			wantErr:     true,
			expectedErr: apperrors.SlugConflict("sudah-dipakai"),
		},
		{
			name: "success - auto-generated slug retried on collision",
			req: &invdto.CreateInvitationRequest{
				Title:        "Pernikahan Budi & Ani",
				EventType:    "wedding",
				EventDate:    time.Now().Add(30 * 24 * time.Hour),
				VenueName:    "Gedung Serbaguna",
				VenueAddress: "Jl. Merdeka No. 1",
			},
			setupMock: func(repo *mock.MockInvitationRepository) {
				gomock.InOrder(
					repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(apperrors.ErrConflict),
					repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil),
				)
			},
			wantErr: false,
		},
		{
			name: "error - repository fails",
			req: &invdto.CreateInvitationRequest{
				Title:        "Pernikahan Budi & Ani",
				EventType:    "wedding",
				EventDate:    time.Now().Add(30 * 24 * time.Hour),
				VenueName:    "Gedung Serbaguna",
				VenueAddress: "Jl. Merdeka No. 1",
			},
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewInvitationUsecase(mockRepo)
			got, err := uc.Create(context.Background(), mockUserID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Create() expected error but got nil")
				}
				if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("Create() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
			}
			if got == nil || got.Title != tt.req.Title {
				t.Errorf("Create() title mismatch, got %v", got)
			}
		})
	}
}

func TestInvitationUsecase_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - draft invitation, no status sync needed",
			setupMock: func(repo *mock.MockInvitationRepository) {
				draft := *mockInvitation
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(&draft, nil)
			},
			wantErr: false,
		},
		{
			name: "success - published but expired, status synced via Update",
			setupMock: func(repo *mock.MockInvitationRepository) {
				expired := *mockInvitation
				expired.IsPublished = true
				expired.EventDate = time.Now().Add(-24 * time.Hour)
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(&expired, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - not found",
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error - owned by a different user returns not-found, not the other user's data",
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockOtherUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(mockRepo)

			userID := mockUserID
			if tt.name == "error - owned by a different user returns not-found, not the other user's data" {
				userID = mockOtherUserID
			}

			uc := usecase.NewInvitationUsecase(mockRepo)
			got, err := uc.GetByID(context.Background(), mockInvitationID, userID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetByID() unexpected error = %v", err)
			}
			if got == nil || got.ID != mockInvitationID {
				t.Errorf("GetByID() got wrong invitation")
			}
		})
	}
}

func TestInvitationUsecase_GetAll(t *testing.T) {
	pg := pagination.Pagination{Page: 1, Limit: 10, Offset: 0}

	tests := []struct {
		name      string
		setupMock func(repo *mock.MockInvitationRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "success - returns list",
			setupMock: func(repo *mock.MockInvitationRepository) {
				inv := *mockInvitation
				repo.EXPECT().FindAllByUserID(gomock.Any(), mockUserID, pg).Return([]*entity.Invitation{&inv}, int64(1), nil)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "error - repository fails",
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().FindAllByUserID(gomock.Any(), mockUserID, pg).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewInvitationUsecase(mockRepo)
			got, err := uc.GetAll(context.Background(), mockUserID, pg)

			if tt.wantErr {
				if err == nil {
					t.Error("GetAll() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetAll() unexpected error = %v", err)
			}
			if len(got.Items) != tt.wantCount {
				t.Errorf("GetAll() items count = %d, want %d", len(got.Items), tt.wantCount)
			}
		})
	}
}

func TestInvitationUsecase_Update(t *testing.T) {
	tests := []struct {
		name        string
		req         *invdto.UpdateInvitationRequest
		setupMock   func(repo *mock.MockInvitationRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - publish sets published_at",
			req:  &invdto.UpdateInvitationRequest{IsPublished: boolPtr(true)},
			setupMock: func(repo *mock.MockInvitationRepository) {
				draft := *mockInvitation
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(&draft, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - not found",
			req:  &invdto.UpdateInvitationRequest{Title: "New Title"},
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			expectedErr: apperrors.InvitationNotFound(mockInvitationID),
		},
		{
			name: "error - repository update fails",
			req:  &invdto.UpdateInvitationRequest{Title: "New Title"},
			setupMock: func(repo *mock.MockInvitationRepository) {
				draft := *mockInvitation
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(&draft, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewInvitationUsecase(mockRepo)
			got, err := uc.Update(context.Background(), mockInvitationID, mockUserID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Update() unexpected error = %v", err)
			}
			if got == nil || !got.IsPublished {
				t.Errorf("Update() expected is_published=true, got %v", got)
			}
			if got.PublishedAt == nil {
				t.Errorf("Update() expected published_at to be set")
			}
		})
	}
}

func TestInvitationUsecase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - deleted",
			setupMock: func(repo *mock.MockInvitationRepository) {
				inv := *mockInvitation
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(&inv, nil)
				repo.EXPECT().Delete(gomock.Any(), mockInvitationID, mockUserID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - not found",
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewInvitationUsecase(mockRepo)
			err := uc.Delete(context.Background(), mockInvitationID, mockUserID)

			if tt.wantErr && err == nil {
				t.Error("Delete() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}
		})
	}
}

func TestInvitationUsecase_GetPublicBySlug(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - published invitation",
			setupMock: func(repo *mock.MockInvitationRepository) {
				published := *mockInvitation
				published.IsPublished = true
				repo.EXPECT().FindBySlug(gomock.Any(), "pernikahan-budi-ani").Return(&published, nil)
			},
			wantErr: false,
		},
		{
			name: "error - draft invitation is not exposed publicly",
			setupMock: func(repo *mock.MockInvitationRepository) {
				draft := *mockInvitation
				repo.EXPECT().FindBySlug(gomock.Any(), "pernikahan-budi-ani").Return(&draft, nil)
			},
			wantErr: true,
		},
		{
			name: "error - slug not found",
			setupMock: func(repo *mock.MockInvitationRepository) {
				repo.EXPECT().FindBySlug(gomock.Any(), "pernikahan-budi-ani").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewInvitationUsecase(mockRepo)
			got, err := uc.GetPublicBySlug(context.Background(), "pernikahan-budi-ani")

			if tt.wantErr {
				if err == nil {
					t.Error("GetPublicBySlug() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPublicBySlug() unexpected error = %v", err)
			}
			if got == nil || got.Title != mockInvitation.Title {
				t.Errorf("GetPublicBySlug() got wrong invitation")
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }
