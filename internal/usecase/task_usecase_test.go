package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
)

var _ repository.TaskRepository = (*taskRepoStubForTask)(nil)
var _ repository.ScheduleRepository = (*scheduleRepoStubForTask)(nil)

type taskRepoStubForTask struct {
	tasks map[string][]domain.Task
}

func (t *taskRepoStubForTask) ListBySchedule(ctx context.Context, scheduleID string) ([]domain.Task, error) {
	return t.tasks[scheduleID], nil
}

func (t *taskRepoStubForTask) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, reason *string) error {
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

func (t *taskRepoStubForTask) CreateTask(ctx context.Context, task domain.Task) (string, error) {
	taskID := "task-" + time.Now().Format("20060102150405")
	task.ID = taskID
	if t.tasks == nil {
		t.tasks = make(map[string][]domain.Task)
	}
	t.tasks[task.ScheduleID] = append(t.tasks[task.ScheduleID], task)
	return taskID, nil
}

type scheduleRepoStubForTask struct {
	schedule domain.Schedule
}

func (s *scheduleRepoStubForTask) ListSchedules(ctx context.Context, caregiverID string, filter repository.ScheduleFilter) ([]domain.ScheduleSummary, error) {
	return nil, nil
}

func (s *scheduleRepoStubForTask) GetSchedule(ctx context.Context, scheduleID string) (domain.Schedule, error) {
	return s.schedule, nil
}

func (s *scheduleRepoStubForTask) GetScheduleForCaregiver(ctx context.Context, scheduleID, caregiverID string) (domain.Schedule, error) {
	if s.schedule.ID != scheduleID || s.schedule.CaregiverID != caregiverID {
		return domain.Schedule{}, domain.ErrNotFound
	}
	return s.schedule, nil
}

func (s *scheduleRepoStubForTask) LogClockIn(ctx context.Context, scheduleID string, event domain.VisitEvent) error {
	return nil
}

func (s *scheduleRepoStubForTask) LogClockOut(ctx context.Context, scheduleID string, event domain.VisitEvent) error {
	return nil
}

func (s *scheduleRepoStubForTask) UpdateStatus(ctx context.Context, scheduleID string, status domain.ScheduleStatus) error {
	return nil
}

func (s *scheduleRepoStubForTask) GetMetrics(ctx context.Context, caregiverID string, day time.Time) (domain.ScheduleMetrics, error) {
	return domain.ScheduleMetrics{}, nil
}

func TestTaskUsecaseUpdateTaskStatus(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{
			"sched-1": {
				{
					ID:         "task-1",
					ScheduleID: "sched-1",
					Title:      "Test Task",
					Status:     domain.TaskStatusPending,
				},
			},
		},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	err := uc.UpdateTaskStatus(context.Background(), "caregiver-1", "sched-1", "task-1", domain.TaskStatusCompleted, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	task := taskRepo.tasks["sched-1"][0]
	if task.Status != domain.TaskStatusCompleted {
		t.Fatalf("expected task status to be completed, got %s", task.Status)
	}
	if task.NotCompletedReason != nil {
		t.Fatalf("expected not completed reason to be nil")
	}
}

func TestTaskUsecaseUpdateTaskStatusNotCompleted(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{
			"sched-1": {
				{
					ID:         "task-1",
					ScheduleID: "sched-1",
					Title:      "Test Task",
					Status:     domain.TaskStatusPending,
				},
			},
		},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	reason := "Client declined service"
	err := uc.UpdateTaskStatus(context.Background(), "caregiver-1", "sched-1", "task-1", domain.TaskStatusNotCompleted, &reason)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	task := taskRepo.tasks["sched-1"][0]
	if task.Status != domain.TaskStatusNotCompleted {
		t.Fatalf("expected task status to be not_completed, got %s", task.Status)
	}
	if task.NotCompletedReason == nil || *task.NotCompletedReason != reason {
		t.Fatalf("expected not completed reason to be %s, got %v", reason, task.NotCompletedReason)
	}
}

func TestTaskUsecaseUpdateTaskStatusNotCompletedWithoutReason(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{
			"sched-1": {
				{
					ID:         "task-1",
					ScheduleID: "sched-1",
					Title:      "Test Task",
					Status:     domain.TaskStatusPending,
				},
			},
		},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	err := uc.UpdateTaskStatus(context.Background(), "caregiver-1", "sched-1", "task-1", domain.TaskStatusNotCompleted, nil)
	if err != domain.ErrValidationFailure {
		t.Fatalf("expected validation failure error, got %v", err)
	}
}

func TestTaskUsecaseUpdateTaskStatusInvalidStatus(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{
			"sched-1": {
				{
					ID:         "task-1",
					ScheduleID: "sched-1",
					Title:      "Test Task",
					Status:     domain.TaskStatusPending,
				},
			},
		},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	err := uc.UpdateTaskStatus(context.Background(), "caregiver-1", "sched-1", "task-1", domain.TaskStatus("invalid"), nil)
	if err != domain.ErrValidationFailure {
		t.Fatalf("expected validation failure error, got %v", err)
	}
}

func TestTaskUsecaseUpdateTaskStatusUnauthorizedSchedule(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-2", // Different caregiver
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{
			"sched-1": {
				{
					ID:         "task-1",
					ScheduleID: "sched-1",
					Title:      "Test Task",
					Status:     domain.TaskStatusPending,
				},
			},
		},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	err := uc.UpdateTaskStatus(context.Background(), "caregiver-1", "sched-1", "task-1", domain.TaskStatusCompleted, nil)
	if err != domain.ErrNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestTaskUsecaseUpdateTaskStatusTaskNotFound(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{
			"sched-1": {
				{
					ID:         "task-1",
					ScheduleID: "sched-1",
					Title:      "Test Task",
					Status:     domain.TaskStatusPending,
				},
			},
		},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	err := uc.UpdateTaskStatus(context.Background(), "caregiver-1", "sched-1", "non-existent-task", domain.TaskStatusCompleted, nil)
	if err != domain.ErrForbidden {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestTaskUsecaseAddTask(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	taskID, err := uc.AddTask(context.Background(), "caregiver-1", "sched-1", domain.Task{
		Title:       "New Task",
		Description: "Task description",
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if taskID == "" {
		t.Fatalf("expected task ID to be returned")
	}
	
	tasks := taskRepo.tasks["sched-1"]
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	
	task := tasks[0]
	if task.Title != "New Task" {
		t.Fatalf("expected task title to be 'New Task', got %s", task.Title)
	}
	if task.Description != "Task description" {
		t.Fatalf("expected task description to be 'Task description', got %s", task.Description)
	}
	if task.SortOrder != 1 {
		t.Fatalf("expected task sort order to be 1, got %d", task.SortOrder)
	}
	if task.Status != domain.TaskStatusPending {
		t.Fatalf("expected task status to be pending, got %s", task.Status)
	}
	if task.ScheduleID != "sched-1" {
		t.Fatalf("expected task schedule ID to be 'sched-1', got %s", task.ScheduleID)
	}
}

func TestTaskUsecaseAddTaskUnauthorizedSchedule(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-2", // Different caregiver
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	_, err := uc.AddTask(context.Background(), "caregiver-1", "sched-1", domain.Task{
		Title: "New Task",
	})
	if err != domain.ErrNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestTaskUsecaseAddTaskEmptyTitle(t *testing.T) {
	scheduleRepo := &scheduleRepoStubForTask{
		schedule: domain.Schedule{
			ID:          "sched-1",
			CaregiverID: "caregiver-1",
			Status:      domain.ScheduleStatusInProgress,
		},
	}
	
	taskRepo := &taskRepoStubForTask{
		tasks: map[string][]domain.Task{},
	}
	
	uc := NewTaskUsecase(taskRepo, scheduleRepo)
	
	_, err := uc.AddTask(context.Background(), "caregiver-1", "sched-1", domain.Task{
		Title: "", // Empty title
	})
	if err != domain.ErrValidationFailure {
		t.Fatalf("expected validation failure error, got %v", err)
	}
}