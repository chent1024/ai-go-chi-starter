package example

import (
	"context"
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
		return Example{}, shared.ErrInternal("repository is not configured")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Example{}, shared.ErrInvalidArgument(
			"name is required",
			shared.WithFieldErrors(shared.RequiredField("name")),
		)
	}
	item := Example{
		ID:   shared.NewID("exm"),
		Name: name,
	}
	return s.repo.Create(ctx, item)
}

func (s *Service) Get(ctx context.Context, id string) (Example, error) {
	if s.repo == nil {
		return Example{}, shared.ErrInternal("repository is not configured")
	}
	if strings.TrimSpace(id) == "" {
		return Example{}, shared.ErrInvalidArgument(
			"id is required",
			shared.WithFieldErrors(shared.RequiredField("id")),
		)
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Example, error) {
	if s.repo == nil {
		return nil, shared.ErrInternal("repository is not configured")
	}
	return s.repo.List(ctx)
}
