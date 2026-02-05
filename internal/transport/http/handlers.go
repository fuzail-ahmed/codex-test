package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/fuzail-ahmed/codex-test/internal/model"
	"github.com/fuzail-ahmed/codex-test/internal/repository"
	"github.com/fuzail-ahmed/codex-test/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/todos", h.handleTodos)
	mux.HandleFunc("/todos/bulk", h.handleBulkCreate)
	mux.HandleFunc("/todos/", h.handleTodoByID)
	mux.HandleFunc("/healthz", h.handleHealth)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreate(w, r)
	case http.MethodGet:
		h.handleList(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.svc.Create(r.Context(), service.CreateTodoInput{Title: req.Title, Description: req.Description})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapTodo(result))
}

func (h *Handler) handleBulkCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Items []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"items"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	inputs := make([]service.CreateTodoInput, 0, len(req.Items))
	for _, item := range req.Items {
		inputs = append(inputs, service.CreateTodoInput{Title: item.Title, Description: item.Description})
	}
	result, err := h.svc.BulkCreate(r.Context(), inputs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapTodos(result))
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	result, err := h.svc.List(r.Context(), repository.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapTodos(result))
}

func (h *Handler) handleTodoByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/todos/")
	if path == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id, err := uuid.Parse(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		result, err := h.svc.Get(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, mapTodo(result))
	case http.MethodPatch:
		var req struct {
			Title       *string `json:"title"`
			Description *string `json:"description"`
			Status      *string `json:"status"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		var status *model.Status
		if req.Status != nil {
			parsed := model.Status(*req.Status)
			status = &parsed
		}
		result, err := h.svc.Update(r.Context(), id, service.UpdateTodoInput{
			Title:       req.Title,
			Description: req.Description,
			Status:      status,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, mapTodo(result))
	case http.MethodDelete:
		deleted, err := h.svc.Delete(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err == nil {
		return errors.New("unexpected extra JSON")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"message": message,
		},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, repository.ErrConflict):
		writeError(w, http.StatusConflict, "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func mapTodo(todo model.Todo) map[string]any {
	return map[string]any{
		"id":          todo.ID.String(),
		"title":       todo.Title,
		"description": todo.Description,
		"status":      todo.Status,
		"created_at":  todo.CreatedAt,
		"updated_at":  todo.UpdatedAt,
	}
}

func mapTodos(todos []model.Todo) map[string]any {
	items := make([]map[string]any, 0, len(todos))
	for _, todo := range todos {
		items = append(items, mapTodo(todo))
	}
	return map[string]any{"items": items}
}

func parseInt(val string, def int) int {
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return n
}
