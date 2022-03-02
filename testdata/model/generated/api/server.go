package api

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/cugu/swagger-go-chi/testdata/model/generated/model"
)

type Service interface {
}

func NewServer(service Service, roleAuth func([]string) func(http.Handler) http.Handler, middlewares ...func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares...)

	s := &server{service}

	return r
}

type server struct {
	service Service
}
