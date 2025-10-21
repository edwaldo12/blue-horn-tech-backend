package domain

import (
	"time"
)

// ScheduleStatus enumerates the various lifecycle states for a caregiver visit.
type ScheduleStatus string

const (
	ScheduleStatusScheduled  ScheduleStatus = "scheduled"
	ScheduleStatusInProgress ScheduleStatus = "in_progress"
	ScheduleStatusCompleted  ScheduleStatus = "completed"
	ScheduleStatusCancelled  ScheduleStatus = "cancelled"
	ScheduleStatusMissed     ScheduleStatus = "missed"
)

// TaskStatus enumerates the completion status for each care activity.
type TaskStatus string

const (
	TaskStatusPending      TaskStatus = "pending"
	TaskStatusCompleted    TaskStatus = "completed"
	TaskStatusNotCompleted TaskStatus = "not_completed"
)

// Caregiver represents the user of the application.
type Caregiver struct {
	ID    string
	Name  string
	Email string
}

// Client represents the care recipient tied to a schedule.
type Client struct {
	ID        string
	FullName  string
	Email     string
	Phone     string
	Address   string
	City      string
	State     string
	Postal    string
	Latitude  float64
	Longitude float64
}

// Task describes a scheduled care activity.
type Task struct {
	ID                 string
	ScheduleID         string
	Title              string
	Description        string
	Status             TaskStatus
	NotCompletedReason *string
	SortOrder          int32
	UpdatedAt          time.Time
}

// Schedule aggregates visit metadata, client details, and tasks.
type Schedule struct {
	ID            string
	CaregiverID   string
	Client        Client
	ServiceName   string
	StartTime     time.Time
	EndTime       time.Time
	Status        ScheduleStatus
	ClockInAt     *time.Time
	ClockInLat    *float64
	ClockInLong   *float64
	ClockOutAt    *time.Time
	ClockOutLat   *float64
	ClockOutLong  *float64
	Notes         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Tasks         []Task
	DurationMins  int
	LocationLabel string
}

// ScheduleSummary is a lightweight projection for listing.
type ScheduleSummary struct {
	ID           string
	CaregiverID  string
	ClientName   string
	ServiceName  string
	StartTime    time.Time
	EndTime      time.Time
	Status       ScheduleStatus
	LocationName string
}

// ScheduleMetrics aggregates dashboard counts.
type ScheduleMetrics struct {
	Total      int `json:"total"`
	Missed     int `json:"missed"`
	Upcoming   int `json:"upcoming"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Cancelled  int `json:"cancelled"`
}

// VisitEvent captures a geolocated clock event.
type VisitEvent struct {
	ScheduleID string
	Latitude   float64
	Longitude  float64
	Timestamp  time.Time
	Notes      *string
}
