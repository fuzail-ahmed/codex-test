package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/fuzail-ahmed/codex-test/internal/model"
	"github.com/fuzail-ahmed/codex-test/internal/repository"
)

type Repo struct {
	mu    sync.RWMutex
	items map[uuid.UUID]model.Todo
}

func New() *Repo {
	return &Repo{items: make(map[uuid.UUID]model.Todo)}
}

func (r *Repo) Create(ctx context.Context, todo model.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[todo.ID]; exists {
		return repository.ErrConflict
	}
	r.items[todo.ID] = todo
	return nil
}

func (r *Repo) CreateBatch(ctx context.Context, todos []model.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, todo := range todos {
		if _, exists := r.items[todo.ID]; exists {
			return repository.ErrConflict
		}
	}
	for _, todo := range todos {
		r.items[todo.ID] = todo
	}
	return nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (model.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	todo, ok := r.items[id]
	if !ok {
		return model.Todo{}, repository.ErrNotFound
	}
	return todo, nil
}

func (r *Repo) List(ctx context.Context, filter repository.ListFilter) ([]model.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Todo, 0, len(r.items))
	for _, todo := range r.items {
		result = append(result, todo)
	}
	return result, nil
}

func (r *Repo) Update(ctx context.Context, todo model.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[todo.ID]; !ok {
		return repository.ErrNotFound
	}
	r.items[todo.ID] = todo
	return nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}
