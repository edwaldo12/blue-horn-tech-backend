package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
)

// ScheduleUsecase orchestrates schedule interactions across repositories.
type ScheduleUsecase struct {
	schedules repository.ScheduleRepository
	tasks     repository.TaskRepository
	now       func() time.Time
}

// NewScheduleUsecase constructs a ScheduleUsecase with sane defaults.
func NewScheduleUsecase(
	schedules repository.ScheduleRepository,
	tasks repository.TaskRepository,
) *ScheduleUsecase {
	return &ScheduleUsecase{
		schedules: schedules,
		tasks:     tasks,
		now:       time.Now,
	}
}

// WithNow allows injecting a deterministic clock for testing.
func (uc *ScheduleUsecase) WithNow(now func() time.Time) {
	if now != nil {
		uc.now = now
	}
}

// ListSchedules returns all schedules for a caregiver given a filter.
func (uc *ScheduleUsecase) ListSchedules(ctx context.Context, caregiverID string, filter repository.ScheduleFilter) ([]domain.ScheduleSummary, error) {
	return uc.schedules.ListSchedules(ctx, caregiverID, filter)
}

// GetSchedule fetches a full schedule including tasks.
func (uc *ScheduleUsecase) GetSchedule(ctx context.Context, scheduleID, caregiverID string) (domain.Schedule, error) {
	schedule, err := uc.schedules.GetScheduleForCaregiver(ctx, scheduleID, caregiverID)
	if err != nil {
		return domain.Schedule{}, err
	}

	tasks, err := uc.tasks.ListBySchedule(ctx, scheduleID)
	if err != nil {
		return domain.Schedule{}, err
	}
	schedule.Tasks = tasks

	return schedule, nil
}

// StartSchedule records a clock-in event and transitions status.
func (uc *ScheduleUsecase) StartSchedule(ctx context.Context, scheduleID, caregiverID string, event domain.VisitEvent) (domain.Schedule, error) {
	schedule, err := uc.schedules.GetScheduleForCaregiver(ctx, scheduleID, caregiverID)
	if err != nil {
		return domain.Schedule{}, err
	}
	if schedule.Status != domain.ScheduleStatusScheduled && schedule.Status != domain.ScheduleStatusMissed {
		return domain.Schedule{}, domain.ErrInvalidStatusTransition
	}

	event.Timestamp = uc.now()
	if err := uc.schedules.LogClockIn(ctx, scheduleID, event); err != nil {
		return domain.Schedule{}, err
	}
	if err := uc.schedules.UpdateStatus(ctx, scheduleID, domain.ScheduleStatusInProgress); err != nil {
		return domain.Schedule{}, err
	}

	return uc.GetSchedule(ctx, scheduleID, caregiverID)
}

// EndSchedule records a clock-out event and transitions status to completed.
func (uc *ScheduleUsecase) EndSchedule(ctx context.Context, scheduleID, caregiverID string, event domain.VisitEvent) (domain.Schedule, error) {
	schedule, err := uc.schedules.GetScheduleForCaregiver(ctx, scheduleID, caregiverID)
	if err != nil {
		return domain.Schedule{}, err
	}
	if schedule.Status != domain.ScheduleStatusInProgress {
		return domain.Schedule{}, domain.ErrInvalidStatusTransition
	}

	event.Timestamp = uc.now()
	if event.Notes != nil && len(*event.Notes) == 0 {
		event.Notes = nil
	}

	if err := uc.schedules.LogClockOut(ctx, scheduleID, event); err != nil {
		return domain.Schedule{}, err
	}
	if err := uc.schedules.UpdateStatus(ctx, scheduleID, domain.ScheduleStatusCompleted); err != nil {
		return domain.Schedule{}, err
	}

	return uc.GetSchedule(ctx, scheduleID, caregiverID)
}

// GetMetrics returns dashboard counts for the given day.
func (uc *ScheduleUsecase) GetMetrics(ctx context.Context, caregiverID string, day time.Time) (domain.ScheduleMetrics, error) {
	return uc.schedules.GetMetrics(ctx, caregiverID, day)
}

// UpdateScheduleStatus allows manual transitions for admin flows.
func (uc *ScheduleUsecase) UpdateScheduleStatus(ctx context.Context, scheduleID string, status domain.ScheduleStatus) error {
	switch status {
	case domain.ScheduleStatusScheduled,
		domain.ScheduleStatusInProgress,
		domain.ScheduleStatusCompleted,
		domain.ScheduleStatusCancelled,
		domain.ScheduleStatusMissed:
	default:
		return domain.ErrValidationFailure
	}
	return uc.schedules.UpdateStatus(ctx, scheduleID, status)
}

// EnsureScheduleOwnership verifies the caregiver relationship.
func (uc *ScheduleUsecase) EnsureScheduleOwnership(ctx context.Context, scheduleID, caregiverID string) error {
	_, err := uc.schedules.GetScheduleForCaregiver(ctx, scheduleID, caregiverID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
