package repository

import (
	"context"
	"taskService/internal/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=task.go -destination=mock_task_repo.go -package=repository
type Task interface {
	Create(ctx context.Context, t *domain.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, int64, error)
	Update(ctx context.Context, t *domain.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}
