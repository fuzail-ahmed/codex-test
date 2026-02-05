package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/fuzail-ahmed/codex-test/internal/model"
	"github.com/fuzail-ahmed/codex-test/internal/repository"
	"github.com/fuzail-ahmed/codex-test/internal/worker"
)

var (
	ErrValidation = errors.New("validation error")
)

type CreateTodoInput struct {
	Title       string
	Description string
}

type UpdateTodoInput struct {
	Title       *string
	Description *string
	Status      *model.Status
}

type Service struct {
	repo        repository.TodoRepository
	workers     int
	now         func() time.Time
	idGenerator func() uuid.UUID
}

func New(repo repository.TodoRepository, workers int) *Service {
	return &Service{
		repo:        repo,
		workers:     workers,
		now:         time.Now,
		idGenerator: uuid.New,
	}
}

func (s *Service) Create(ctx context.Context, input CreateTodoInput) (model.Todo, error) {
	if err := validateCreate(input); err != nil {
		return model.Todo{}, err
	}
	now := s.now()
	todo := model.Todo{
		ID:          s.idGenerator(),
		Title:       input.Title,
		Description: input.Description,
		Status:      model.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, todo); err != nil {
		return model.Todo{}, err
	}
	return todo, nil
}

func (s *Service) BulkCreate(ctx context.Context, inputs []CreateTodoInput) ([]model.Todo, error) {
	if len(inputs) == 0 {
		return nil, wrapValidation("items must not be empty")
	}

	pool := worker.Pool[CreateTodoInput, model.Todo]{
		Workers: s.workers,
		Work: func(ctx context.Context, input CreateTodoInput) (model.Todo, error) {
			if err := validateCreate(input); err != nil {
				return model.Todo{}, err
			}
			now := s.now()
			return model.Todo{
				ID:          s.idGenerator(),
				Title:       input.Title,
				Description: input.Description,
				Status:      model.StatusPending,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	todos, err := pool.Run(ctx, inputs)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateBatch(ctx, todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (model.Todo, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, filter repository.ListFilter) ([]model.Todo, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTodoInput) (model.Todo, error) {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Todo{}, err
	}

	if input.Title != nil {
		if err := validateTitle(*input.Title); err != nil {
			return model.Todo{}, err
		}
		existing.Title = *input.Title
	}
	if input.Description != nil {
		if err := validateDescription(*input.Description); err != nil {
			return model.Todo{}, err
		}
		existing.Description = *input.Description
	}
	if input.Status != nil {
		if err := validateStatus(*input.Status); err != nil {
			return model.Todo{}, err
		}
		existing.Status = *input.Status
	}
	if input.Title == nil && input.Description == nil && input.Status == nil {
		return model.Todo{}, wrapValidation("no fields to update")
	}
	existing.UpdatedAt = s.now()

	if err := s.repo.Update(ctx, existing); err != nil {
		return model.Todo{}, err
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.Delete(ctx, id)
}
