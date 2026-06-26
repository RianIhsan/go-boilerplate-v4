package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	guestentity "github.com/RianIhsan/go-boilerplate-v4/internal/domain/guest/entity"
	invitationentity "github.com/RianIhsan/go-boilerplate-v4/internal/domain/invitation/entity"
	rsvpdto "github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/entity"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/rsvp/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	"github.com/RianIhsan/go-boilerplate-v4/pkg/pagination"
	"github.com/golang/mock/gomock"
)

var (
	mockUserID       = "user-id-1"
	mockInvitationID = "invitation-id-1"
	mockGuestID      = "guest-id-1"
	mockToken        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	mockSlug         = "pernikahan-budi-ani"
	mockInvitation   = &invitationentity.Invitation{
		ID:          mockInvitationID,
		UserID:      mockUserID,
		Title:       "Pernikahan Budi & Ani",
		Slug:        mockSlug,
		EventDate:   time.Now().Add(30 * 24 * time.Hour),
		IsPublished: true,
	}
	mockGuest = &guestentity.Guest{
		ID:           mockGuestID,
		InvitationID: mockInvitationID,
		Name:         "Tamu Satu",
		UniqueToken:  mockToken,
	}
)

func TestRSVPUsecase_Submit(t *testing.T) {
	tests := []struct {
		name      string
		req       *rsvpdto.SubmitRSVPRequest
		setupMock func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository, guestRepo *mock.MockGuestRepository)
		wantErr   bool
	}{
		{
			name: "success - anonymous rsvp without guest token",
			req:  &rsvpdto.SubmitRSVPRequest{Name: "Tamu Tanpa Undangan", Status: "attending", AttendeeCount: 2},
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository, guestRepo *mock.MockGuestRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockSlug).Return(mockInvitation, nil)
				rsvpRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - first rsvp via personal guest token creates a new record",
			req:  &rsvpdto.SubmitRSVPRequest{GuestToken: mockToken, Status: "attending"},
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository, guestRepo *mock.MockGuestRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockSlug).Return(mockInvitation, nil)
				guestRepo.EXPECT().FindByToken(gomock.Any(), mockToken).Return(mockGuest, nil)
				rsvpRepo.EXPECT().FindByGuestID(gomock.Any(), mockGuestID).Return(nil, errors.New("not found"))
				rsvpRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - resubmitting via the same guest token updates the existing record",
			req:  &rsvpdto.SubmitRSVPRequest{GuestToken: mockToken, Status: "not_attending"},
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository, guestRepo *mock.MockGuestRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockSlug).Return(mockInvitation, nil)
				guestRepo.EXPECT().FindByToken(gomock.Any(), mockToken).Return(mockGuest, nil)
				existing := &entity.RSVP{ID: "rsvp-id-1", InvitationID: mockInvitationID, GuestID: &mockGuestID, Status: entity.RSVPStatusAttending}
				rsvpRepo.EXPECT().FindByGuestID(gomock.Any(), mockGuestID).Return(existing, nil)
				rsvpRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - slug not found or not published",
			req:  &rsvpdto.SubmitRSVPRequest{Name: "Tamu", Status: "attending"},
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository, guestRepo *mock.MockGuestRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockSlug).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error - guest token does not belong to this invitation",
			req:  &rsvpdto.SubmitRSVPRequest{GuestToken: mockToken, Status: "attending"},
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository, guestRepo *mock.MockGuestRepository) {
				invRepo.EXPECT().FindBySlug(gomock.Any(), mockSlug).Return(mockInvitation, nil)
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

			rsvpRepo := mock.NewMockRSVPRepository(ctrl)
			invRepo := mock.NewMockInvitationRepository(ctrl)
			guestRepo := mock.NewMockGuestRepository(ctrl)
			tt.setupMock(rsvpRepo, invRepo, guestRepo)

			uc := usecase.NewRSVPUsecase(rsvpRepo, invRepo, guestRepo)
			got, err := uc.Submit(context.Background(), mockSlug, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Submit() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Submit() unexpected error = %v", err)
			}
			if got == nil || got.InvitationID != mockInvitationID {
				t.Errorf("Submit() got wrong rsvp")
			}
		})
	}
}

func TestRSVPUsecase_GetAll(t *testing.T) {
	pg := pagination.Pagination{Page: 1, Limit: 10, Offset: 0}

	tests := []struct {
		name      string
		setupMock func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository)
		wantErr   bool
	}{
		{
			name: "success - returns rsvps for owned invitation",
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(mockInvitation, nil)
				rsvp := &entity.RSVP{ID: "rsvp-id-1", InvitationID: mockInvitationID, Status: entity.RSVPStatusAttending}
				rsvpRepo.EXPECT().FindAllByInvitationID(gomock.Any(), mockInvitationID, pg).Return([]*entity.RSVP{rsvp}, int64(1), nil)
			},
			wantErr: false,
		},
		{
			name: "error - invitation not owned by caller",
			setupMock: func(rsvpRepo *mock.MockRSVPRepository, invRepo *mock.MockInvitationRepository) {
				invRepo.EXPECT().FindByID(gomock.Any(), mockInvitationID, mockUserID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			rsvpRepo := mock.NewMockRSVPRepository(ctrl)
			invRepo := mock.NewMockInvitationRepository(ctrl)
			guestRepo := mock.NewMockGuestRepository(ctrl)
			tt.setupMock(rsvpRepo, invRepo)

			uc := usecase.NewRSVPUsecase(rsvpRepo, invRepo, guestRepo)
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
