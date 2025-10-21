package repository

import (
	"context"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
)

// CaregiverLogRepository defines the interface for caregiver log data operations.
type CaregiverLogRepository interface {
	// CreateLog creates a new caregiver log entry.
	CreateLog(ctx context.Context, log domain.CaregiverLog) (string, error)
	
	// GetLogsByCaregiver retrieves logs for a caregiver with pagination.
	GetLogsByCaregiver(ctx context.Context, caregiverID string, filter CaregiverLogFilter) ([]domain.CaregiverLogSummary, error)
	
	// GetTodayLogs retrieves today's logs for a caregiver.
	GetTodayLogs(ctx context.Context, caregiverID string) ([]domain.CaregiverLog, error)
	
	// GetLastClockIn retrieves the last clock-in log for a caregiver.
	GetLastClockIn(ctx context.Context, caregiverID string) (*domain.CaregiverLog, error)
	
	// GetLastClockOut retrieves the last clock-out log for a caregiver.
	GetLastClockOut(ctx context.Context, caregiverID string) (*domain.CaregiverLog, error)
	
	// HasClockedInToday checks if caregiver has clocked in today.
	HasClockedInToday(ctx context.Context, caregiverID string) (bool, error)
	
	// HasClockedOutToday checks if caregiver has clocked out today.
	HasClockedOutToday(ctx context.Context, caregiverID string) (bool, error)
}

// CaregiverLogFilter defines filtering options for caregiver log queries.
type CaregiverLogFilter struct {
	LogType   *domain.LogType
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}