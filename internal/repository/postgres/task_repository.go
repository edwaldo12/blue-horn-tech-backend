package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

// TaskRepository implements repository.TaskRepository with Postgres.
type TaskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository constructs the repository.
func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

type taskRow struct {
	ID                 string         `db:"id"`
	ScheduleID         string         `db:"schedule_id"`
	Title              string         `db:"title"`
	Description        sql.NullString `db:"description"`
	Status             string         `db:"status"`
	NotCompletedReason sql.NullString `db:"not_completed_reason"`
	SortOrder          int32          `db:"sort_order"`
	UpdatedAt          sql.NullTime   `db:"updated_at"`
}

// ListBySchedule returns all tasks belonging to the schedule ordered by sort order.
func (r *TaskRepository) ListBySchedule(ctx context.Context, scheduleID string) ([]domain.Task, error) {
	query := `
		SELECT id,
		       schedule_id,
		       title,
		       description,
		       status,
		       not_completed_reason,
		       sort_order,
		       updated_at
		FROM schedule_tasks
		WHERE schedule_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`
	var rows []taskRow
	if err := r.db.SelectContext(ctx, &rows, query, scheduleID); err != nil {
		return nil, err
	}

	result := make([]domain.Task, len(rows))
	for i, row := range rows {
		var description string
		if row.Description.Valid {
			description = row.Description.String
		}
		var reason *string
		if row.NotCompletedReason.Valid {
			val := row.NotCompletedReason.String
			reason = &val
		}
		var updatedAt time.Time
		if row.UpdatedAt.Valid {
			updatedAt = row.UpdatedAt.Time
		}

		result[i] = domain.Task{
			ID:                 row.ID,
			ScheduleID:         row.ScheduleID,
			Title:              row.Title,
			Description:        description,
			Status:             domain.TaskStatus(row.Status),
			NotCompletedReason: reason,
			SortOrder:          row.SortOrder,
			UpdatedAt:          updatedAt,
		}
	}
	return result, nil
}

// UpdateTaskStatus changes the task status and optional reason.
func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, reason *string) error {
	query := `
		UPDATE schedule_tasks
		SET status = $2,
		    not_completed_reason = $3,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, taskID, status, reason)
	return err
}

// CreateTask inserts a new task record.
func (r *TaskRepository) CreateTask(ctx context.Context, task domain.Task) (string, error) {
	query := `
		INSERT INTO schedule_tasks (id, schedule_id, title, description, status, not_completed_reason, sort_order)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, COALESCE($6, 0))
		RETURNING id
	`
	var id string
	err := r.db.QueryRowContext(ctx, query,
		task.ScheduleID,
		task.Title,
		nullable(task.Description),
		task.Status,
		task.NotCompletedReason,
		task.SortOrder,
	).Scan(&id)
	return id, err
}

func nullable(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
