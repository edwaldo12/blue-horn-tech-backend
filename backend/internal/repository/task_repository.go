package repository

import (
	"context"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
)

// TaskRepository defines persistence operations for care activities.
type TaskRepository interface {
	ListBySchedule(ctx context.Context, scheduleID string) ([]domain.Task, error)
	UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, reason *string) error
	CreateTask(ctx context.Context, task domain.Task) (string, error)
}
