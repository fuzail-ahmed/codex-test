package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/fuzail-ahmed/codex-test/internal/model"
)

type ListFilter struct {
	Limit  int
	Offset int
}

type TodoRepository interface {
	Create(ctx context.Context, todo model.Todo) error
	CreateBatch(ctx context.Context, todos []model.Todo) error
	Get(ctx context.Context, id uuid.UUID) (model.Todo, error)
	List(ctx context.Context, filter ListFilter) ([]model.Todo, error)
	Update(ctx context.Context, todo model.Todo) error
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
}
