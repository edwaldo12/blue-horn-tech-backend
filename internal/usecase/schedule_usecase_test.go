package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
)

var _ repository.ScheduleRepository = (*scheduleRepoStub)(nil)
var _ repository.TaskRepository = (*taskRepoStub)(nil)

type scheduleRepoStub struct {
	schedule domain.Schedule
}

func (s *scheduleRepoStub) ListSchedules(ctx context.Context, caregiverID string, filter repository.ScheduleFilter) ([]domain.ScheduleSummary, error) {
	return nil, nil
}

func (s *scheduleRepoStub) GetSchedule(ctx context.Context, scheduleID string) (domain.Schedule, error) {
	return s.schedule, nil
}

func (s *scheduleRepoStub) GetScheduleForCaregiver(ctx context.Context, scheduleID, caregiverID string) (domain.Schedule, error) {
	if s.schedule.ID != scheduleID || s.schedule.CaregiverID != caregiverID {
		return domain.Schedule{}, domain.ErrNotFound
	}
	return s.schedule, nil
}

func (s *scheduleRepoStub) LogClockIn(ctx context.Context, scheduleID string, event domain.VisitEvent) error {
	if s.schedule.ID != scheduleID {
		return domain.ErrNotFound
	}
	s.schedule.ClockInAt = &event.Timestamp
	s.schedule.ClockInLat = &event.Latitude
	s.schedule.ClockInLong = &event.Longitude
	return nil
}

func (s *scheduleRepoStub) LogClockOut(ctx context.Context, scheduleID string, event domain.VisitEvent) error {
	if s.schedule.ID != scheduleID {
		return domain.ErrNotFound
	}
	s.schedule.ClockOutAt = &event.Timestamp
	s.schedule.ClockOutLat = &event.Latitude
	s.schedule.ClockOutLong = &event.Longitude
	return nil
}

func (s *scheduleRepoStub) UpdateStatus(ctx context.Context, scheduleID string, status domain.ScheduleStatus) error {
	if s.schedule.ID != scheduleID {
		return domain.ErrNotFound
	}
	s.schedule.Status = status
	return nil
}

func (s *scheduleRepoStub) GetMetrics(ctx context.Context, caregiverID string, day time.Time) (domain.ScheduleMetrics, error) {
	return domain.ScheduleMetrics{}, nil
}

type taskRepoStub struct {
	tasks map[string][]domain.Task
}

func (t *taskRepoStub) ListBySchedule(ctx context.Context, scheduleID string) ([]domain.Task, error) {
	return t.tasks[scheduleID], nil
}

func (t *taskRepoStub) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, reason *string) error {
	for scheduleID, list := range t.tasks {
		for i := range list {
			if list[i].ID == taskID {
				list[i].Status = status
				list[i].NotCompletedReason = reason
				t.tasks[scheduleID][i] = list[i]
				return nil
			}
		}
	}
	return domain.ErrNotFound
}

func (t *taskRepoStub) CreateTask(ctx context.Context, task domain.Task) (string, error) {
	return "task-123", nil
}

func TestScheduleUsecaseStartSchedule(t *testing.T) {
	now := time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC)
	schedule := domain.Schedule{
		ID:          "sched-1",
		CaregiverID: "cg-1",
		Status:      domain.ScheduleStatusScheduled,
	}
	repo := &scheduleRepoStub{schedule: schedule}
	tasks := &taskRepoStub{tasks: map[string][]domain.Task{
		"sched-1": {
			{ID: "task-1", Title: "Task", Status: domain.TaskStatusPending},
		},
	}}

	uc := NewScheduleUsecase(repo, tasks)
	uc.WithNow(func() time.Time { return now })

	result, err := uc.StartSchedule(context.Background(), "sched-1", "cg-1", domain.VisitEvent{Latitude: 1.23, Longitude: 4.56})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.schedule.Status != domain.ScheduleStatusInProgress {
		t.Fatalf("expected status in_progress, got %s", repo.schedule.Status)
	}
	if repo.schedule.ClockInAt == nil || !repo.schedule.ClockInAt.Equal(now) {
		t.Fatalf("expected clock-in timestamp to be %v", now)
	}
	if len(result.Tasks) != 1 {
		t.Fatalf("expected tasks to be loaded")
	}
}

func TestScheduleUsecaseStartScheduleInvalidTransition(t *testing.T) {
	repo := &scheduleRepoStub{schedule: domain.Schedule{
		ID:          "sched-1",
		CaregiverID: "cg-1",
		Status:      domain.ScheduleStatusInProgress,
	}}
	tasks := &taskRepoStub{tasks: map[string][]domain.Task{"sched-1": {}}}

	uc := NewScheduleUsecase(repo, tasks)
	_, err := uc.StartSchedule(context.Background(), "sched-1", "cg-1", domain.VisitEvent{Latitude: 1, Longitude: 1})
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}

func TestScheduleUsecaseEndSchedule(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	clockIn := now.Add(-time.Hour)
	repo := &scheduleRepoStub{schedule: domain.Schedule{
		ID:          "sched-1",
		CaregiverID: "cg-1",
		Status:      domain.ScheduleStatusInProgress,
		ClockInAt:   &clockIn,
	}}
	tasks := &taskRepoStub{tasks: map[string][]domain.Task{"sched-1": {}}}

	uc := NewScheduleUsecase(repo, tasks)
	uc.WithNow(func() time.Time { return now })

	_, err := uc.EndSchedule(context.Background(), "sched-1", "cg-1", domain.VisitEvent{Latitude: 1, Longitude: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.schedule.Status != domain.ScheduleStatusCompleted {
		t.Fatalf("expected status completed, got %s", repo.schedule.Status)
	}
	if repo.schedule.ClockOutAt == nil || !repo.schedule.ClockOutAt.Equal(now) {
		t.Fatalf("expected clock-out timestamp to be %v", now)
	}
}

func TestTaskUsecaseRequiresReason(t *testing.T) {
	repo := &scheduleRepoStub{schedule: domain.Schedule{ID: "sched-1", CaregiverID: "cg-1", Status: domain.ScheduleStatusInProgress}}
	tasks := &taskRepoStub{tasks: map[string][]domain.Task{
		"sched-1": {
			{ID: "task-1", ScheduleID: "sched-1", Status: domain.TaskStatusPending},
		},
	}}
	NewScheduleUsecase(repo, tasks)
	taskUC := NewTaskUsecase(tasks, repo)

	err := taskUC.UpdateTaskStatus(context.Background(), "cg-1", "sched-1", "task-1", domain.TaskStatusNotCompleted, nil)
	if !errors.Is(err, domain.ErrValidationFailure) {
		t.Fatalf("expected validation failure, got %v", err)
	}

	reason := "Client declined"
	err = taskUC.UpdateTaskStatus(context.Background(), "cg-1", "sched-1", "task-1", domain.TaskStatusNotCompleted, &reason)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tasks.tasks["sched-1"][0].NotCompletedReason == nil || *tasks.tasks["sched-1"][0].NotCompletedReason != reason {
		t.Fatalf("expected reason to be persisted")
	}
}
