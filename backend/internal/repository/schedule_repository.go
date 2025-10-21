package repository

import (
	"context"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
)

// ScheduleFilter captures optional parameters for listing schedules.
type ScheduleFilter struct {
	Status     []domain.ScheduleStatus
	Date       *time.Time
	TodayOnly  bool
	WithClient bool
	// Pagination parameters
	Limit      int
	Offset     int
}

// ScheduleRepository defines the persistence contract for schedule operations.
type ScheduleRepository interface {
	ListSchedules(ctx context.Context, caregiverID string, filter ScheduleFilter) ([]domain.ScheduleSummary, error)
	GetSchedule(ctx context.Context, scheduleID string) (domain.Schedule, error)
	GetScheduleForCaregiver(ctx context.Context, scheduleID, caregiverID string) (domain.Schedule, error)
	LogClockIn(ctx context.Context, scheduleID string, event domain.VisitEvent) error
	LogClockOut(ctx context.Context, scheduleID string, event domain.VisitEvent) error
	UpdateStatus(ctx context.Context, scheduleID string, status domain.ScheduleStatus) error
	GetMetrics(ctx context.Context, caregiverID string, day time.Time) (domain.ScheduleMetrics, error)
}
