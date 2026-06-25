package handler_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/handler"
	"github.com/RianIhsan/go-boilerplate-v4/internal/mock"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
)

var mockFileMetaResponse = &dto.FileMetadataResponse{
	FileType: "rpm",
	Metadata: map[string]any{
		"name":    "epel-release",
		"version": "7-5",
		"vendor":  "Fedora Project",
	},
}

func buildMultipartRequest(t *testing.T, content []byte) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if content != nil {
		fw, err := w.CreateFormFile("file", "upload.bin")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/metadata", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestFileMetaHandler_ParseMetadata(t *testing.T) {
	tests := []struct {
		name           string
		body           *http.Request
		setupMock      func(uc *mock.MockFileMetaUsecase)
		expectedStatus int
	}{
		{
			name: "success - 200",
			body: buildMultipartRequest(t, []byte("fake rpm bytes")),
			setupMock: func(uc *mock.MockFileMetaUsecase) {
				uc.EXPECT().ParseMetadata(gomock.Any(), gomock.Any()).Return(mockFileMetaResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - missing file field",
			body:           buildMultipartRequest(t, nil),
			setupMock:      func(uc *mock.MockFileMetaUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "error - usecase rejects unsupported type",
			body: buildMultipartRequest(t, []byte("garbage")),
			setupMock: func(uc *mock.MockFileMetaUsecase) {
				uc.EXPECT().ParseMetadata(gomock.Any(), gomock.Any()).Return(nil, apperrors.UnsupportedFileType())
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "error - usecase internal failure",
			body: buildMultipartRequest(t, []byte("anything")),
			setupMock: func(uc *mock.MockFileMetaUsecase) {
				uc.EXPECT().ParseMetadata(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrInternalServer)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock.NewMockFileMetaUsecase(ctrl)
			tt.setupMock(mockUC)
			h := handler.NewFileMetaHandler(mockUC)

			rr := httptest.NewRecorder()
			h.ParseMetadata(rr, tt.body)

			if rr.Code != tt.expectedStatus {
				t.Errorf("ParseMetadata() status = %d, want %d, body: %s", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestFileMetaHandler_ParseMetadata_TooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockFileMetaUsecase(ctrl)
	h := handler.NewFileMetaHandler(mockUC)

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		fw, err := mw.CreateFormFile("file", "big.bin")
		if err == nil {
			_, _ = io.CopyN(fw, zeroReader{}, constants.MaxUploadFileBytes+1)
		}
		_ = mw.Close()
		_ = pw.Close()
	}()

	req := httptest.NewRequest(http.MethodPost, "/files/metadata", pr)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()

	h.ParseMetadata(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("ParseMetadata() status = %d, want %d, body: %s", rr.Code, http.StatusRequestEntityTooLarge, rr.Body.String())
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func TestRegisterRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockFileMetaUsecase(ctrl)
	mockUC.EXPECT().ParseMetadata(gomock.Any(), gomock.Any()).Return(mockFileMetaResponse, nil)

	h := handler.NewFileMetaHandler(mockUC)

	noopRateLimit := func(next http.Handler) http.Handler { return next }

	r := chi.NewRouter()
	handler.RegisterRoutes(r, h, noopRateLimit)

	req := buildMultipartRequest(t, []byte("fake rpm bytes"))
	req.URL.Path = "/files/metadata"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /files/metadata: status = %d, want %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
