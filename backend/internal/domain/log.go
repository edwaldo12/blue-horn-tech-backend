package domain

import "time"

// LogType represents the type of caregiver log entry.
type LogType string

const (
	LogTypeClockIn  LogType = "clock_in"
	LogTypeClockOut LogType = "clock_out"
)

// CaregiverLog models a caregiver attendance log entry.
type CaregiverLog struct {
	ID          string
	CaregiverID string
	LogType     LogType
	Latitude    float64
	Longitude   float64
	Timestamp   time.Time
	Notes       *string
	CreatedAt   time.Time
}

// CaregiverLogSummary is a lightweight projection for listing.
type CaregiverLogSummary struct {
	ID          string
	CaregiverID string
	LogType     LogType
	Timestamp   time.Time
}

// RequestLog models an HTTP request/response log entry persisted for auditing.
type RequestLog struct {
	ID        string
	Method    string
	Path      string
	Query     string
	Status    int
	Latency   time.Duration
	IP        string
	UserAgent string
	CreatedAt time.Time
}
