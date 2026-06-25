package handler

import (
	"errors"
	"net/http"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/usecase"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
	apperrors "github.com/RianIhsan/go-boilerplate-v4/internal/shared/errors"
	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/response"
)

type FileMetaHandler struct {
	fileMetaUsecase usecase.FileMetaUsecase
}

func NewFileMetaHandler(fileMetaUsecase usecase.FileMetaUsecase) *FileMetaHandler {
	return &FileMetaHandler{fileMetaUsecase: fileMetaUsecase}
}

func (h *FileMetaHandler) ParseMetadata(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxUploadFileBytes)

	if err := r.ParseMultipartForm(constants.MaxUploadFileBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			response.Error(w, r, apperrors.FileTooLarge())
			return
		}
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.Error(w, r, apperrors.ErrBadRequest)
		return
	}
	defer file.Close()

	result, err := h.fileMetaUsecase.ParseMetadata(r.Context(), file)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, http.StatusOK, "file metadata parsed successfully", result)
}
