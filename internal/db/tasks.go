package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xdars/grpc-tasks/internal/domain"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Add(ctx context.Context, t *domain.Task) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tasks (id, user_id, title, description, status) VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.UserID, t.Title, t.Description, t.Status,
	)
	return err
}

func (r *TaskRepository) Get(ctx context.Context, id string) (*domain.Task, bool, error) {
	t := &domain.Task{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, description FROM tasks WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return t, true, nil
}

func (r *TaskRepository) GetByUser(ctx context.Context, userID string) ([]*domain.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, description, status FROM tasks WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func (r *TaskRepository) Update(ctx context.Context, t *domain.Task) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET title=$1, description=$2, status=$3, updated_at=NOW() WHERE id=$4`,
		t.Title, t.Description, t.Status, t.ID,
	)
	return err
}
