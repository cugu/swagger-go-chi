package api

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/cugu/swagger-go-chi/testdata/formData/generated/model"
)

type Service interface {
	UploadFile(context.Context, []*multipart.FileHeader, []string) error
}

func NewServer(service Service, roleAuth func([]string) func(http.Handler) http.Handler, middlewares ...func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares...)

	s := &server{service}

	r.With(roleAuth([]string{"uploadSystemData"})).Put("/file", s.uploadFileHandler)
	return r
}

type server struct {
	service Service
}

func (s *server) uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	const MaxFileSize = 32 << 20 // maximum file size of about 32 MB
	if r.ContentLength > MaxFileSize {
		JSONErrorStatus(w, http.StatusExpectationFailed, errors.New("request too large"))
		return
	}
	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		JSONError(w, err)
		return
	}
	uploadP := r.MultipartForm.File["upload"]

	metadataP := r.MultipartForm.Value["metadata"]

	response(w, nil, s.service.UploadFile(r.Context(), uploadP, metadataP))
}
