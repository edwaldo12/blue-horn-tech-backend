package usecase

import (
	"context"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
)

// CaregiverAttendanceUsecase orchestrates caregiver attendance operations.
type CaregiverAttendanceUsecase struct {
	caregiverLogs repository.CaregiverLogRepository
	now           func() time.Time
}

// NewCaregiverAttendanceUsecase creates a CaregiverAttendanceUsecase instance.
func NewCaregiverAttendanceUsecase(caregiverLogs repository.CaregiverLogRepository) *CaregiverAttendanceUsecase {
	return &CaregiverAttendanceUsecase{
		caregiverLogs: caregiverLogs,
		now:           time.Now,
	}
}

// WithNow allows injecting a deterministic clock for testing.
func (uc *CaregiverAttendanceUsecase) WithNow(now func() time.Time) {
	if now != nil {
		uc.now = now
	}
}

// ClockIn records a clock-in event for the caregiver.
func (uc *CaregiverAttendanceUsecase) ClockIn(ctx context.Context, caregiverID string, latitude, longitude float64, notes *string) (domain.CaregiverLog, error) {
	// Check if already clocked in today
	hasClockedIn, err := uc.caregiverLogs.HasClockedInToday(ctx, caregiverID)
	if err != nil {
		return domain.CaregiverLog{}, err
	}
	if hasClockedIn {
		return domain.CaregiverLog{}, domain.ErrValidationFailure
	}

	log := domain.CaregiverLog{
		CaregiverID: caregiverID,
		LogType:     domain.LogTypeClockIn,
		Latitude:    latitude,
		Longitude:   longitude,
		Timestamp:   uc.now(),
		Notes:       notes,
	}

	id, err := uc.caregiverLogs.CreateLog(ctx, log)
	if err != nil {
		return domain.CaregiverLog{}, err
	}

	log.ID = id
	return log, nil
}

// ClockOut records a clock-out event for the caregiver.
func (uc *CaregiverAttendanceUsecase) ClockOut(ctx context.Context, caregiverID string, latitude, longitude float64, notes *string) (domain.CaregiverLog, error) {
	// Check if clocked in today
	hasClockedIn, err := uc.caregiverLogs.HasClockedInToday(ctx, caregiverID)
	if err != nil {
		return domain.CaregiverLog{}, err
	}
	if !hasClockedIn {
		return domain.CaregiverLog{}, domain.ErrValidationFailure
	}

	// Check if already clocked out today
	hasClockedOut, err := uc.caregiverLogs.HasClockedOutToday(ctx, caregiverID)
	if err != nil {
		return domain.CaregiverLog{}, err
	}
	if hasClockedOut {
		return domain.CaregiverLog{}, domain.ErrValidationFailure
	}

	log := domain.CaregiverLog{
		CaregiverID: caregiverID,
		LogType:     domain.LogTypeClockOut,
		Latitude:    latitude,
		Longitude:   longitude,
		Timestamp:   uc.now(),
		Notes:       notes,
	}

	id, err := uc.caregiverLogs.CreateLog(ctx, log)
	if err != nil {
		return domain.CaregiverLog{}, err
	}

	log.ID = id
	return log, nil
}

// GetTodayStatus returns today's attendance status for the caregiver.
func (uc *CaregiverAttendanceUsecase) GetTodayStatus(ctx context.Context, caregiverID string) (TodayAttendanceStatus, error) {
	todayLogs, err := uc.caregiverLogs.GetTodayLogs(ctx, caregiverID)
	if err != nil {
		return TodayAttendanceStatus{}, err
	}

	status := TodayAttendanceStatus{
		CaregiverID: caregiverID,
		Date:        time.Date(uc.now().Year(), uc.now().Month(), uc.now().Day(), 0, 0, 0, 0, uc.now().Location()),
		HasClockedIn: false,
		HasClockedOut: false,
	}

	for _, log := range todayLogs {
		if log.LogType == domain.LogTypeClockIn {
			status.HasClockedIn = true
			status.ClockInAt = &log.Timestamp
			status.ClockInLat = &log.Latitude
			status.ClockInLong = &log.Longitude
		} else if log.LogType == domain.LogTypeClockOut {
			status.HasClockedOut = true
			status.ClockOutAt = &log.Timestamp
			status.ClockOutLat = &log.Latitude
			status.ClockOutLong = &log.Longitude
		}
	}

	return status, nil
}

// GetAttendanceHistory retrieves attendance history for a caregiver.
func (uc *CaregiverAttendanceUsecase) GetAttendanceHistory(ctx context.Context, caregiverID string, filter repository.CaregiverLogFilter) ([]domain.CaregiverLogSummary, error) {
	// Set default limit if not provided
	if filter.Limit == 0 {
		filter.Limit = 100 // Default to last 100 logs
	}

	return uc.caregiverLogs.GetLogsByCaregiver(ctx, caregiverID, filter)
}

// TodayAttendanceStatus represents the attendance status for today.
type TodayAttendanceStatus struct {
	CaregiverID   string
	Date          time.Time
	HasClockedIn  bool
	HasClockedOut bool
	ClockInAt     *time.Time
	ClockInLat    *float64
	ClockInLong   *float64
	ClockOutAt    *time.Time
	ClockOutLat   *float64
	ClockOutLong  *float64
}