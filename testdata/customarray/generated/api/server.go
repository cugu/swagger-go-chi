package api

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/cugu/swagger-go-chi/testdata/customarray/generated/model"
)

type Service interface {
	CreateUserBatch(context.Context, *model.UserArray) error
}

func NewServer(service Service, roleAuth func([]string) func(http.Handler) http.Handler, middlewares ...func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares...)

	s := &server{service}

	r.With(roleAuth([]string{""})).Post("/users", s.createUserBatchHandler)
	return r
}

type server struct {
	service Service
}

func (s *server) createUserBatchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, err)
		return
	}

	if validateSchema(body, model.UserArraySchema, w) {
		return
	}

	var usersP *model.UserArray
	if err := parseBody(body, &usersP); err != nil {
		JSONError(w, err)
		return
	}

	response(w, nil, s.service.CreateUserBatch(r.Context(), usersP))
}
