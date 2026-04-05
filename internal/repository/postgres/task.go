package postgres

import (
	"context"
	"fmt"
	"strings"
	"taskService/internal/domain"
	"taskService/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepo(pool *pgxpool.Pool) repository.Task {
	return &taskRepo{pool: pool}
}

func (r *taskRepo) Create(ctx context.Context, t *domain.Task) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tasks (id, title, description, status, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.Title, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *taskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t := &domain.Task{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, description, status, created_at, updated_at, deleted_at 
		 FROM tasks WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	return t, err
}

func (r *taskRepo) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, int64, error) {
	var where []string
	var args []any
	idx := 1

	where = append(where, "deleted_at IS NULL")
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, *f.Status)
		idx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereClause), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf(`SELECT id, title, description, status, created_at, updated_at, deleted_at 
	                  FROM tasks WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, idx, idx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *taskRepo) Update(ctx context.Context, t *domain.Task) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET title=$2, description=$3, status=$4, updated_at=$5 
		 WHERE id=$1 AND deleted_at IS NULL`,
		t.ID, t.Title, t.Description, t.Status, t.UpdatedAt)
	return err
}

func (r *taskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET deleted_at=$2, updated_at=$2 WHERE id=$1 AND deleted_at IS NULL`,
		id, time.Now())
	return err
}
