package usecase

import (
	"context"
	"io"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/dto"
)

//go:generate mockgen -source=filemeta_usecase.go -destination=../../../mock/mock_filemeta_usecase.go -package=mock

type FileMetaUsecase interface {
	ParseMetadata(ctx context.Context, file io.Reader) (*dto.FileMetadataResponse, error)
}
