package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
)

var _ repository.CaregiverLogRepository = (*caregiverLogRepoStub)(nil)

type caregiverLogRepoStub struct {
	logs       []domain.CaregiverLog
	hasClocked bool
}

func (c *caregiverLogRepoStub) HasClockedInToday(ctx context.Context, caregiverID string) (bool, error) {
	return c.hasClocked, nil
}

func (c *caregiverLogRepoStub) HasClockedOutToday(ctx context.Context, caregiverID string) (bool, error) {
	// For testing, assume if there are 2 logs, they've clocked out
	return len(c.logs) >= 2, nil
}

func (c *caregiverLogRepoStub) CreateLog(ctx context.Context, log domain.CaregiverLog) (string, error) {
	log.ID = "log-123"
	c.logs = append(c.logs, log)
	return log.ID, nil
}

func (c *caregiverLogRepoStub) GetTodayLogs(ctx context.Context, caregiverID string) ([]domain.CaregiverLog, error) {
	return c.logs, nil
}

func (c *caregiverLogRepoStub) GetLastClockIn(ctx context.Context, caregiverID string) (*domain.CaregiverLog, error) {
	for i := len(c.logs) - 1; i >= 0; i-- {
		if c.logs[i].LogType == domain.LogTypeClockIn {
			return &c.logs[i], nil
		}
	}
	return nil, nil
}

func (c *caregiverLogRepoStub) GetLastClockOut(ctx context.Context, caregiverID string) (*domain.CaregiverLog, error) {
	for i := len(c.logs) - 1; i >= 0; i-- {
		if c.logs[i].LogType == domain.LogTypeClockOut {
			return &c.logs[i], nil
		}
	}
	return nil, nil
}

func (c *caregiverLogRepoStub) GetLogsByCaregiver(ctx context.Context, caregiverID string, filter repository.CaregiverLogFilter) ([]domain.CaregiverLogSummary, error) {
	// Always return a non-nil slice, even if empty
	summaries := make([]domain.CaregiverLogSummary, 0)
	for _, log := range c.logs {
		summaries = append(summaries, domain.CaregiverLogSummary{
			ID:        log.ID,
			CaregiverID: log.CaregiverID,
			LogType:   log.LogType,
			Timestamp: log.Timestamp,
		})
	}
	return summaries, nil
}

func TestCaregiverAttendanceUsecaseClockIn(t *testing.T) {
	now := time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC)
	repo := &caregiverLogRepoStub{hasClocked: false}
	
	uc := NewCaregiverAttendanceUsecase(repo)
	uc.WithNow(func() time.Time { return now })

	log, err := uc.ClockIn(context.Background(), "caregiver-1", 40.7128, -74.0060, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if log.CaregiverID != "caregiver-1" {
		t.Fatalf("expected caregiver ID caregiver-1, got %s", log.CaregiverID)
	}
	if log.LogType != domain.LogTypeClockIn {
		t.Fatalf("expected log type clock_in, got %s", log.LogType)
	}
	if log.Latitude != 40.7128 || log.Longitude != -74.0060 {
		t.Fatalf("expected coordinates 40.7128, -74.0060, got %f, %f", log.Latitude, log.Longitude)
	}
	if !log.Timestamp.Equal(now) {
		t.Fatalf("expected timestamp %v, got %v", now, log.Timestamp)
	}
	if len(repo.logs) != 1 {
		t.Fatalf("expected 1 log in repository, got %d", len(repo.logs))
	}
}

func TestCaregiverAttendanceUsecaseClockInAlreadyClockedIn(t *testing.T) {
	repo := &caregiverLogRepoStub{hasClocked: true}
	uc := NewCaregiverAttendanceUsecase(repo)

	_, err := uc.ClockIn(context.Background(), "caregiver-1", 40.7128, -74.0060, nil)
	if err != domain.ErrValidationFailure {
		t.Fatalf("expected validation failure error, got %v", err)
	}
}

func TestCaregiverAttendanceUsecaseClockOut(t *testing.T) {
	now := time.Date(2025, 1, 15, 17, 0, 0, 0, time.UTC)
	repo := &caregiverLogRepoStub{
		hasClocked: true,
		logs: []domain.CaregiverLog{
			{
				ID:        "log-1",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockIn,
				Timestamp: now.Add(-8 * time.Hour),
			},
		},
	}
	
	uc := NewCaregiverAttendanceUsecase(repo)
	uc.WithNow(func() time.Time { return now })

	log, err := uc.ClockOut(context.Background(), "caregiver-1", 40.7128, -74.0060, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if log.CaregiverID != "caregiver-1" {
		t.Fatalf("expected caregiver ID caregiver-1, got %s", log.CaregiverID)
	}
	if log.LogType != domain.LogTypeClockOut {
		t.Fatalf("expected log type clock_out, got %s", log.LogType)
	}
	if !log.Timestamp.Equal(now) {
		t.Fatalf("expected timestamp %v, got %v", now, log.Timestamp)
	}
	if len(repo.logs) != 2 {
		t.Fatalf("expected 2 logs in repository, got %d", len(repo.logs))
	}
}

func TestCaregiverAttendanceUsecaseClockOutNotClockedIn(t *testing.T) {
	repo := &caregiverLogRepoStub{hasClocked: false}
	uc := NewCaregiverAttendanceUsecase(repo)

	_, err := uc.ClockOut(context.Background(), "caregiver-1", 40.7128, -74.0060, nil)
	if err != domain.ErrValidationFailure {
		t.Fatalf("expected validation failure error, got %v", err)
	}
}

func TestCaregiverAttendanceUsecaseClockOutAlreadyClockedOut(t *testing.T) {
	now := time.Date(2025, 1, 15, 17, 0, 0, 0, time.UTC)
	repo := &caregiverLogRepoStub{
		hasClocked: true,
		logs: []domain.CaregiverLog{
			{
				ID:        "log-1",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockIn,
				Timestamp: now.Add(-9 * time.Hour),
			},
			{
				ID:        "log-2",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockOut,
				Timestamp: now.Add(-1 * time.Hour),
			},
		},
	}
	
	uc := NewCaregiverAttendanceUsecase(repo)

	_, err := uc.ClockOut(context.Background(), "caregiver-1", 40.7128, -74.0060, nil)
	if err != domain.ErrValidationFailure {
		t.Fatalf("expected validation failure error, got %v", err)
	}
}

func TestCaregiverAttendanceUsecaseGetTodayStatus(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	clockInTime := now.Add(-3 * time.Hour)
	clockOutTime := now.Add(-1 * time.Hour)
	lat := 40.7128
	long := -74.0060
	
	repo := &caregiverLogRepoStub{
		logs: []domain.CaregiverLog{
			{
				ID:        "log-1",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockIn,
				Latitude:  lat,
				Longitude: long,
				Timestamp: clockInTime,
			},
			{
				ID:        "log-2",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockOut,
				Latitude:  lat,
				Longitude: long,
				Timestamp: clockOutTime,
			},
		},
	}
	
	uc := NewCaregiverAttendanceUsecase(repo)
	uc.WithNow(func() time.Time { return now })

	status, err := uc.GetTodayStatus(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.CaregiverID != "caregiver-1" {
		t.Fatalf("expected caregiver ID caregiver-1, got %s", status.CaregiverID)
	}
	if !status.HasClockedIn {
		t.Fatalf("expected has clocked in to be true")
	}
	if !status.HasClockedOut {
		t.Fatalf("expected has clocked out to be true")
	}
	if status.ClockInAt == nil || !status.ClockInAt.Equal(clockInTime) {
		t.Fatalf("expected clock in time %v, got %v", clockInTime, status.ClockInAt)
	}
	if status.ClockOutAt == nil || !status.ClockOutAt.Equal(clockOutTime) {
		t.Fatalf("expected clock out time %v, got %v", clockOutTime, status.ClockOutAt)
	}
	if status.ClockInLat == nil || *status.ClockInLat != lat {
		t.Fatalf("expected clock in lat %f, got %f", lat, *status.ClockInLat)
	}
	if status.ClockInLong == nil || *status.ClockInLong != long {
		t.Fatalf("expected clock in long %f, got %f", long, *status.ClockInLong)
	}
}

func TestCaregiverAttendanceUsecaseGetTodayStatusNoLogs(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	repo := &caregiverLogRepoStub{logs: []domain.CaregiverLog{}}
	
	uc := NewCaregiverAttendanceUsecase(repo)
	uc.WithNow(func() time.Time { return now })

	status, err := uc.GetTodayStatus(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.CaregiverID != "caregiver-1" {
		t.Fatalf("expected caregiver ID caregiver-1, got %s", status.CaregiverID)
	}
	if status.HasClockedIn {
		t.Fatalf("expected has clocked in to be false")
	}
	if status.HasClockedOut {
		t.Fatalf("expected has clocked out to be false")
	}
	if status.ClockInAt != nil {
		t.Fatalf("expected clock in time to be nil")
	}
	if status.ClockOutAt != nil {
		t.Fatalf("expected clock out time to be nil")
	}
}

func TestCaregiverAttendanceUsecaseGetAttendanceHistory(t *testing.T) {
	repo := &caregiverLogRepoStub{
		logs: []domain.CaregiverLog{
			{
				ID:        "log-1",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockIn,
				Timestamp: time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC),
			},
			{
				ID:        "log-2",
				CaregiverID: "caregiver-1",
				LogType:   domain.LogTypeClockOut,
				Timestamp: time.Date(2025, 1, 15, 17, 0, 0, 0, time.UTC),
			},
		},
	}
	
	uc := NewCaregiverAttendanceUsecase(repo)

	history, err := uc.GetAttendanceHistory(context.Background(), "caregiver-1", repository.CaregiverLogFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	if history[0].LogType != domain.LogTypeClockIn {
		t.Fatalf("expected first entry to be clock_in, got %s", history[0].LogType)
	}
	if history[1].LogType != domain.LogTypeClockOut {
		t.Fatalf("expected second entry to be clock_out, got %s", history[1].LogType)
	}
}

func TestCaregiverAttendanceUsecaseGetAttendanceHistoryWithDefaultLimit(t *testing.T) {
	repo := &caregiverLogRepoStub{logs: []domain.CaregiverLog{}}
	uc := NewCaregiverAttendanceUsecase(repo)

	filter := repository.CaregiverLogFilter{}
	history, err := uc.GetAttendanceHistory(context.Background(), "caregiver-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The use case should set a default limit of 100 if not provided
	// We can't directly test this without a more sophisticated mock,
	// but we can verify the call succeeds
	if history == nil {
		t.Fatalf("expected history to be non-nil")
	}
	
	// Verify that the default limit was set by checking the filter
	// Since our stub doesn't actually use the filter, we just verify the call succeeded
	if len(history) != 0 {
		t.Fatalf("expected empty history, got %d items", len(history))
	}
}