package usecase

import (
	"context"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
)

// TaskUsecase orchestrates task level operations.
type TaskUsecase struct {
	tasks     repository.TaskRepository
	schedules repository.ScheduleRepository
}

// NewTaskUsecase creates a TaskUsecase instance.
func NewTaskUsecase(tasks repository.TaskRepository, schedules repository.ScheduleRepository) *TaskUsecase {
	return &TaskUsecase{
		tasks:     tasks,
		schedules: schedules,
	}
}

// UpdateTaskStatus validates caregiver ownership before updating.
func (uc *TaskUsecase) UpdateTaskStatus(ctx context.Context, caregiverID, scheduleID, taskID string, status domain.TaskStatus, reason *string) error {
	switch status {
	case domain.TaskStatusCompleted, domain.TaskStatusNotCompleted:
	default:
		return domain.ErrValidationFailure
	}

	if err := uc.validateOwnership(ctx, caregiverID, scheduleID, taskID); err != nil {
		return err
	}

	if status == domain.TaskStatusNotCompleted {
		if reason == nil || len(*reason) == 0 {
			return domain.ErrValidationFailure
		}
	} else {
		reason = nil
	}

	return uc.tasks.UpdateTaskStatus(ctx, taskID, status, reason)
}

// AddTask appends a new task to the schedule.
func (uc *TaskUsecase) AddTask(ctx context.Context, caregiverID, scheduleID string, task domain.Task) (string, error) {
	if err := uc.validateOwnership(ctx, caregiverID, scheduleID, ""); err != nil {
		return "", err
	}
	task.ScheduleID = scheduleID
	if task.Title == "" {
		return "", domain.ErrValidationFailure
	}
	task.Status = domain.TaskStatusPending
	return uc.tasks.CreateTask(ctx, task)
}

func (uc *TaskUsecase) validateOwnership(ctx context.Context, caregiverID, scheduleID, taskID string) error {
	schedule, err := uc.schedules.GetScheduleForCaregiver(ctx, scheduleID, caregiverID)
	if err != nil {
		return err
	}
	if taskID == "" {
		return nil
	}
	tasks, err := uc.tasks.ListBySchedule(ctx, schedule.ID)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == taskID {
			return nil
		}
	}
	return domain.ErrForbidden
}
