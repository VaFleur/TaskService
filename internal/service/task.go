package service

import (
	"context"
	"fmt"
	"taskService/internal/domain"
	"taskService/internal/repository"
	"time"

	"github.com/google/uuid"
)

type TaskService struct {
	repo repository.Task
}

func NewTaskService(repo repository.Task) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, t *domain.Task) error {
	if t.Title == "" {
		return fmt.Errorf("title is required")
	}
	t.ID = uuid.New()
	now := time.Now()
	t.CreatedAt, t.UpdatedAt = now, now
	if t.Status == "" {
		t.Status = "pending"
	}
	return s.repo.Create(ctx, t)
}

func (s *TaskService) List(ctx context.Context, f domain.TaskFilter) (domain.PaginatedTasks, error) {
	tasks, total, err := s.repo.List(ctx, f)
	if err != nil {
		return domain.PaginatedTasks{}, err
	}
	return domain.PaginatedTasks{
		Data:       tasks,
		Pagination: domain.Pagination{Total: total, Limit: f.Limit, Offset: f.Offset},
	}, nil
}

func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) Update(ctx context.Context, id uuid.UUID, t *domain.Task) error {
	t.ID = id
	t.UpdatedAt = time.Now()
	return s.repo.Update(ctx, t)
}

func (s *TaskService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
