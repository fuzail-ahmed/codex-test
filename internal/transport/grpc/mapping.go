package grpcserver

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/fuzail-ahmed/codex-test/internal/model"
	todov1 "github.com/fuzail-ahmed/codex-test/shared/gen/todo/v1"
)

func mapTodo(todo model.Todo) *todov1.Todo {
	return &todov1.Todo{
		Id:            todo.ID.String(),
		Title:         todo.Title,
		Description:   todo.Description,
		Status:        mapStatusToProto(todo.Status),
		CreatedAtUnix: todo.CreatedAt.Unix(),
		UpdatedAtUnix: todo.UpdatedAt.Unix(),
	}
}

func mapTodos(todos []model.Todo) []*todov1.Todo {
	items := make([]*todov1.Todo, 0, len(todos))
	for _, todo := range todos {
		items = append(items, mapTodo(todo))
	}
	return items
}

func mapStatusToProto(status model.Status) todov1.Status {
	switch status {
	case model.StatusPending:
		return todov1.Status_STATUS_PENDING
	case model.StatusDone:
		return todov1.Status_STATUS_DONE
	default:
		return todov1.Status_STATUS_UNSPECIFIED
	}
}

func mapStatus(status todov1.Status) model.Status {
	switch status {
	case todov1.Status_STATUS_PENDING:
		return model.StatusPending
	case todov1.Status_STATUS_DONE:
		return model.StatusDone
	default:
		return model.StatusPending
	}
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid id")
	}
	return id, nil
}
