package service

import (
	"fmt"
	"strings"

	"github.com/fuzail-ahmed/codex-test/internal/model"
)

func validateCreate(input CreateTodoInput) error {
	if err := validateTitle(input.Title); err != nil {
		return err
	}
	if err := validateDescription(input.Description); err != nil {
		return err
	}
	return nil
}

func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return wrapValidation("title is required")
	}
	if len(title) > 200 {
		return wrapValidation("title too long")
	}
	return nil
}

func validateDescription(description string) error {
	if len(description) > 2000 {
		return wrapValidation("description too long")
	}
	return nil
}

func validateStatus(status model.Status) error {
	switch status {
	case model.StatusPending, model.StatusDone:
		return nil
	default:
		return wrapValidation("invalid status")
	}
}

func wrapValidation(message string) error {
	return fmt.Errorf("%w: %s", ErrValidation, message)
}
