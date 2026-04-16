package example

import (
	"context"
	"net/http"
	"strings"

	"ai-go-chi-starter/internal/service/shared"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Example, error) {
	if s.repo == nil {
		return Example{}, shared.NewError("INTERNAL", "repository is not configured", http.StatusInternalServerError)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Example{}, shared.NewError("INVALID_ARGUMENT", "name is required", http.StatusBadRequest)
	}
	item := Example{
		ID:   shared.NewID("exm"),
		Name: name,
	}
	return s.repo.Create(ctx, item)
}

func (s *Service) Get(ctx context.Context, id string) (Example, error) {
	if s.repo == nil {
		return Example{}, shared.NewError("INTERNAL", "repository is not configured", http.StatusInternalServerError)
	}
	if strings.TrimSpace(id) == "" {
		return Example{}, shared.NewError("INVALID_ARGUMENT", "id is required", http.StatusBadRequest)
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Example, error) {
	if s.repo == nil {
		return nil, shared.NewError("INTERNAL", "repository is not configured", http.StatusInternalServerError)
	}
	return s.repo.List(ctx)
}
