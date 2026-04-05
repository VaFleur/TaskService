package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"taskService/internal/domain"
	"taskService/internal/service"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// withCtxTimeout оборачивает вызов функции с таймаутом 5 секунд
func withCtxTimeout(ctx context.Context, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return fn(ctx)
}

// @Summary Create task
// @Accept json
// @Produce json
// @Param task body domain.Task true "Task data"
// @Success 201 {object} domain.Task
// @Router /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := withCtxTimeout(r.Context(), func(ctx context.Context) error {
		return h.svc.Create(ctx, &task)
	}); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// @Summary List tasks
// @Produce json
// @Param status query string false "Filter by status (pending, in_progress, done)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.PaginatedTasks
// @Router /tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	f := domain.TaskFilter{Limit: 20, Offset: 0}
	if s := r.URL.Query().Get("status"); s != "" {
		f.Status = &s
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			f.Limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			f.Offset = v
		}
	}

	var res domain.PaginatedTasks
	if err := withCtxTimeout(r.Context(), func(ctx context.Context) error {
		var err error
		res, err = h.svc.List(ctx, f)
		return err
	}); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// @Summary Get task by ID
// @Produce json
// @Param id path string true "Task UUID"
// @Success 200 {object} domain.Task
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Task not found"
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid task ID"}`, http.StatusBadRequest)
		return
	}

	var task *domain.Task
	if err := withCtxTimeout(r.Context(), func(ctx context.Context) error {
		task, err = h.svc.GetByID(ctx, id)
		return err
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// @Summary Update task
// @Accept json
// @Produce json
// @Param id path string true "Task UUID"
// @Param task body domain.Task true "Task data"
// @Success 200 {object} domain.Task
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Task not found"
// @Router /tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid task ID"}`, http.StatusBadRequest)
		return
	}

	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := withCtxTimeout(r.Context(), func(ctx context.Context) error {
		return h.svc.Update(ctx, id, &task)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленные данные
	task.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// @Summary Delete task (soft delete)
// @Produce json
// @Param id path string true "Task UUID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Task not found"
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid task ID"}`, http.StatusBadRequest)
		return
	}

	if err := withCtxTimeout(r.Context(), func(ctx context.Context) error {
		return h.svc.Delete(ctx, id)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
