package usecase_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
)

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func assertTempDirEmpty(t *testing.T) {
	t.Helper()

	entries, err := os.ReadDir(constants.UploadTempDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("failed to read upload temp dir: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("upload temp dir not cleaned up, found: %v", names)
	}
}

func TestFileMetaUsecase_ParseMetadata(t *testing.T) {
	tests := []struct {
		name            string
		file            io.Reader
		wantErr         bool
		expectedErrCode string
	}{
		{
			name:            "error - unsupported file type",
			file:            bytes.NewReader([]byte("just some random garbage bytes")),
			wantErr:         true,
			expectedErrCode: "UNSUPPORTED_FILE_TYPE",
		},
		{
			name: "error - msi explicitly rejected, not silently mis-parsed",
			file: bytes.NewReader(append(
				[]byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1},
				make([]byte, 512)...,
			)),
			wantErr:         true,
			expectedErrCode: "UNSUPPORTED_FILE_TYPE",
		},
		{
			name:            "error - file too large",
			file:            io.LimitReader(zeroReader{}, constants.MaxUploadFileBytes+1),
			wantErr:         true,
			expectedErrCode: "FILE_TOO_LARGE",
		},
	}

	uc := usecase.NewFileMetaUsecase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.ParseMetadata(context.Background(), tt.file)

			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseMetadata() error = nil, want error")
				}
				appErr, ok := err.(*apperrors.AppError)
				if !ok {
					t.Fatalf("ParseMetadata() error is not *apperrors.AppError: %v", err)
				}
				if appErr.ErrCode != tt.expectedErrCode {
					t.Errorf("ParseMetadata() ErrCode = %q, want %q", appErr.ErrCode, tt.expectedErrCode)
				}
			} else if err != nil {
				t.Fatalf("ParseMetadata() unexpected error = %v", err)
			}

			assertTempDirEmpty(t)
		})
	}
}

func TestFileMetaUsecase_ParseMetadata_Success(t *testing.T) {
	uc := usecase.NewFileMetaUsecase()

	rpmBytes, err := os.ReadFile("parser/testdata/sample.rpm")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := uc.ParseMetadata(context.Background(), bytes.NewReader(rpmBytes))
	if err != nil {
		t.Fatalf("ParseMetadata() unexpected error = %v", err)
	}

	if resp.FileType != "rpm" {
		t.Errorf("FileType = %q, want %q", resp.FileType, "rpm")
	}
	meta, ok := resp.Metadata.(map[string]any)
	if !ok {
		t.Fatalf("Metadata is %T, want map[string]any", resp.Metadata)
	}
	if meta["name"] != "epel-release" {
		t.Errorf("Metadata[name] = %q, want %q", meta["name"], "epel-release")
	}

	if resp.PackageName != "epel-release" {
		t.Errorf("PackageName = %q, want %q", resp.PackageName, "epel-release")
	}
	if resp.Version != "7-5" {
		t.Errorf("Version = %q, want %q", resp.Version, "7-5")
	}
	if resp.Publisher != "Fedora Project" {
		t.Errorf("Publisher = %q, want %q", resp.Publisher, "Fedora Project")
	}

	assertTempDirEmpty(t)
}
