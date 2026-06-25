package usecase

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/dto"
	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/usecase/parser"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
)

const parseTimeout = 10 * time.Second

type fileMetaUsecase struct{}

func NewFileMetaUsecase() FileMetaUsecase {
	return &fileMetaUsecase{}
}

func (u *fileMetaUsecase) ParseMetadata(ctx context.Context, file io.Reader) (*dto.FileMetadataResponse, error) {
	tmpPath, err := saveTemp(file)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpPath)

	return parseWithTimeout(ctx, tmpPath)
}

func saveTemp(file io.Reader) (string, error) {
	if err := os.MkdirAll(constants.UploadTempDir, 0750); err != nil {
		return "", apperrors.Wrap(apperrors.ErrInternalServer, err)
	}

	tmp, err := os.CreateTemp(constants.UploadTempDir, "upload-*")
	if err != nil {
		return "", apperrors.Wrap(apperrors.ErrInternalServer, err)
	}
	defer tmp.Close()

	limited := io.LimitReader(file, constants.MaxUploadFileBytes+1)
	written, err := io.Copy(tmp, limited)
	if err != nil {
		os.Remove(tmp.Name())
		return "", apperrors.Wrap(apperrors.ErrInternalServer, err)
	}
	if written > constants.MaxUploadFileBytes {
		os.Remove(tmp.Name())
		return "", apperrors.FileTooLarge()
	}

	return tmp.Name(), nil
}

func parseWithTimeout(ctx context.Context, path string) (*dto.FileMetadataResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, parseTimeout)
	defer cancel()

	type result struct {
		resp *dto.FileMetadataResponse
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				ch <- result{nil, apperrors.FileParseFailed()}
			}
		}()
		resp, err := detectAndParse(path)
		ch <- result{resp, err}
	}()

	select {
	case <-ctx.Done():
		return nil, apperrors.Wrap(apperrors.FileParseFailed(), ctx.Err())
	case res := <-ch:
		return res.resp, res.err
	}
}

func detectAndParse(path string) (*dto.FileMetadataResponse, error) {
	fileType, err := parser.Detect(path)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.FileParseFailed(), err)
	}

	var meta any
	switch fileType {
	case parser.FileTypeAPK:
		meta, err = parser.ParseAPK(path)
	case parser.FileTypeEXE:
		meta, err = parser.ParseEXE(path)
	case parser.FileTypeDEB:
		meta, err = parser.ParseDEB(path)
	case parser.FileTypeRPM:
		meta, err = parser.ParseRPM(path)
	case parser.FileTypeDMG:
		meta, err = parser.ParseDMG(path)
	default:
		return nil, apperrors.UnsupportedFileType()
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.FileParseFailed(), err)
	}

	n := normalize(fileType, meta)
	return &dto.FileMetadataResponse{
		FileType:    string(fileType),
		PackageName: n.PackageName,
		Version:     n.Version,
		Publisher:   n.Publisher,
		Metadata:    meta,
	}, nil
}
