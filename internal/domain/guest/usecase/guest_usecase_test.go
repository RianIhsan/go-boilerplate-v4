package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	guestdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/entity"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/usecase"
	invitationentity "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/entity"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/golang/mock/gomock"
)

var (
	mockUserID       = "user-id-1"
	mockInvitationID = "invitation-id-1"
	mockGuestID      = "guest-id-1"
	mockToken        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	mockInvitation   = &invitationentity.Invitation{
		ID:          mockInvitationID,
		UserID:      mockUserID,
		Title:       "Pernikahan Budi & Ani",
		Slug:        "pernikahan-budi-ani",
		EventType:   "wedding",
		EventDate:   time.Now().Add(30 * 24 * time.Hour),
		IsPublished: true,
	}
	mockGuest = &entity.Guest{
		ID:           mockGuestID,
		InvitationID: mockInvitationID,
		Name:         "Tamu Satu",
		UniqueToken:  mockToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
)

func TestGuestUsecase_Create(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - guest created under owned invitation",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(mockInvitation, nil)
				guestRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - invitation not owned by caller, guest is never created",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error - repository fails",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(mockInvitation, nil)
				guestRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			guestRepo := mock.NewMockGuestRepository(ctrl)
			invRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(guestRepo, invRepo)

			uc := usecase.NewGuestUsecase(guestRepo, invRepo)
			req := &guestdto.CreateGuestRequest{Name: "Tamu Satu"}
			got, err := uc.Create(context.Background(), mockUserID, mockInvitationID, req)

			if tt.wantErr {
				if err == nil {
					t.Error("Create() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
			}
			if got == nil || got.UniqueToken == "" {
				t.Errorf("Create() expected a generated token, got %v", got)
			}
		})
	}
}

func TestGuestUsecase_GetAll(t *testing.T) {
	pg := pagination.Pagination{Page: 1, Limit: 10, Offset: 0}

	tests := []struct {
		name      string
		setupMock func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - returns guests for owned invitation",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(mockInvitation, nil)
				guestRepo.EXPECT().FindAllByInvitationID(gomock.Any(), mockInvitationID, pg).Return([]*entity.Guest{mockGuest}, int64(1), nil)
			},
			wantErr: false,
		},
		{
			name: "error - invitation not owned by caller",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			guestRepo := mock.NewMockGuestRepository(ctrl)
			invRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(guestRepo, invRepo)

			uc := usecase.NewGuestUsecase(guestRepo, invRepo)
			got, err := uc.GetAll(context.Background(), mockUserID, mockInvitationID, pg)

			if tt.wantErr {
				if err == nil {
					t.Error("GetAll() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetAll() unexpected error = %v", err)
			}
			if len(got.Items) != 1 {
				t.Errorf("GetAll() items count = %d, want 1", len(got.Items))
			}
		})
	}
}

func TestGuestUsecase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - deleted",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(mockInvitation, nil)
				guestRepo.EXPECT().Delete(gomock.Any(), mockGuestID, mockInvitationID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - invitation not owned by caller, guest is never deleted",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error - guest not found under invitation",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(mockInvitation, nil)
				guestRepo.EXPECT().Delete(gomock.Any(), mockGuestID, mockInvitationID).Return(errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			guestRepo := mock.NewMockGuestRepository(ctrl)
			invRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(guestRepo, invRepo)

			uc := usecase.NewGuestUsecase(guestRepo, invRepo)
			err := uc.Delete(context.Background(), mockUserID, mockInvitationID, mockGuestID)

			if tt.wantErr && err == nil {
				t.Error("Delete() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}
		})
	}
}

func TestGuestUsecase_GetPublicByToken(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - token belongs to the invitation in the slug",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockInvitation.Slug).Return(mockInvitation, nil)
				guestRepo.EXPECT().FindByToken(gomock.Any(), mockToken).Return(mockGuest, nil)
			},
			wantErr: false,
		},
		{
			name: "error - slug not found or not published",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockInvitation.Slug).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error - token does not belong to this invitation",
			setupMock: func(guestRepo *mock.MockGuestRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockInvitation.Slug).Return(mockInvitation, nil)
				otherGuest := *mockGuest
				otherGuest.InvitationID = "some-other-invitation"
				guestRepo.EXPECT().FindByToken(gomock.Any(), mockToken).Return(&otherGuest, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			guestRepo := mock.NewMockGuestRepository(ctrl)
			invRepo := mock.NewMockInvitationRepository(ctrl)
			tt.setupMock(guestRepo, invRepo)

			uc := usecase.NewGuestUsecase(guestRepo, invRepo)
			got, err := uc.GetPublicByToken(context.Background(), mockInvitation.Slug, mockToken)

			if tt.wantErr {
				if err == nil {
					t.Error("GetPublicByToken() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPublicByToken() unexpected error = %v", err)
			}
			if got == nil || got.GuestName != mockGuest.Name {
				t.Errorf("GetPublicByToken() got wrong guest")
			}
		})
	}
}
