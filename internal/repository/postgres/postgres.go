package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/fuzail-ahmed/codex-test/internal/model"
	"github.com/fuzail-ahmed/codex-test/internal/repository"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, todo model.Todo) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO todos (id, title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, todo.ID, todo.Title, todo.Description, string(todo.Status), todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrConflict
		}
		return err
	}
	return nil
}

func (r *Repo) CreateBatch(ctx context.Context, todos []model.Todo) error {
	if len(todos) == 0 {
		return nil
	}
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		query, args := buildBatchInsert(todos)
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			if isUniqueViolation(err) {
				return repository.ErrConflict
			}
			return err
		}
		return nil
	})
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (model.Todo, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, status, created_at, updated_at
		FROM todos
		WHERE id = $1
	`, id)

	var todo model.Todo
	var status string
	if err := row.Scan(&todo.ID, &todo.Title, &todo.Description, &status, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return model.Todo{}, repository.ErrNotFound
		}
		return model.Todo{}, err
	}
	todo.Status = model.Status(status)
	return todo, nil
}

func (r *Repo) List(ctx context.Context, filter repository.ListFilter) ([]model.Todo, error) {
	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, status, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []model.Todo{}
	for rows.Next() {
		var todo model.Todo
		var status string
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &status, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			return nil, err
		}
		todo.Status = model.Status(status)
		result = append(result, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repo) Update(ctx context.Context, todo model.Todo) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE todos
		SET title = $2, description = $3, status = $4, updated_at = $5
		WHERE id = $1
	`, todo.ID, todo.Title, todo.Description, string(todo.Status), todo.UpdatedAt)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func buildBatchInsert(todos []model.Todo) (string, []any) {
	var sb strings.Builder
	args := make([]any, 0, len(todos)*6)
	fmt.Fprint(&sb, "INSERT INTO todos (id, title, description, status, created_at, updated_at) VALUES ")
	for i, todo := range todos {
		if i > 0 {
			sb.WriteString(",")
		}
		idx := i*6 + 1
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)", idx, idx+1, idx+2, idx+3, idx+4, idx+5))
		args = append(args, todo.ID, todo.Title, todo.Description, string(todo.Status), todo.CreatedAt, todo.UpdatedAt)
	}
	return sb.String(), args
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint")
}
