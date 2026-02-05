package service

import (
	"context"
	"testing"

	"github.com/fuzail-ahmed/codex-test/internal/repository"
	"github.com/fuzail-ahmed/codex-test/internal/repository/memory"
)

func TestBulkCreate_AllOrNothing(t *testing.T) {
	repo := memory.New()
	svc := New(repo, 4)

	_, err := svc.BulkCreate(context.Background(), []CreateTodoInput{
		{Title: "valid", Description: "ok"},
		{Title: "", Description: "invalid"},
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	items, err := repo.List(context.Background(), repository.ListFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no items, got %d", len(items))
	}
}
