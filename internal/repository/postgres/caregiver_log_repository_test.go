package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

func TestCaregiverLogRepositoryCreateLog(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	log := domain.CaregiverLog{
		CaregiverID: "caregiver-1",
		LogType:     domain.LogTypeClockIn,
		Latitude:    40.7128,
		Longitude:   -74.0060,
		Timestamp:   now,
		Notes:       nil,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO logs_caregivers (caregiver_id, log_type, latitude, longitude, timestamp, notes)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `)).
		WithArgs(log.CaregiverID, log.LogType, log.Latitude, log.Longitude, log.Timestamp, log.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("log-123"))

	id, err := repo.CreateLog(context.Background(), log)
	if err != nil {
		t.Fatalf("CreateLog error: %v", err)
	}
	if id != "log-123" {
		t.Fatalf("expected log ID log-123, got %s", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryCreateLogWithNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	notes := "Running late"
	log := domain.CaregiverLog{
		CaregiverID: "caregiver-1",
		LogType:     domain.LogTypeClockOut,
		Latitude:    40.7128,
		Longitude:   -74.0060,
		Timestamp:   now,
		Notes:       &notes,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO logs_caregivers (caregiver_id, log_type, latitude, longitude, timestamp, notes)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `)).
		WithArgs(log.CaregiverID, log.LogType, log.Latitude, log.Longitude, log.Timestamp, log.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("log-456"))

	id, err := repo.CreateLog(context.Background(), log)
	if err != nil {
		t.Fatalf("CreateLog error: %v", err)
	}
	if id != "log-456" {
		t.Fatalf("expected log ID log-456, got %s", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryGetLogsByCaregiver(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "caregiver_id", "log_type", "timestamp"}).
		AddRow("log-1", "caregiver-1", "clock_in", now).
		AddRow("log-2", "caregiver-1", "clock_out", now.Add(time.Hour))

	mock.ExpectQuery("SELECT id, caregiver_id, log_type, timestamp FROM logs_caregivers WHERE caregiver_id = \\$1 ORDER BY timestamp DESC").
		WithArgs("caregiver-1").
		WillReturnRows(rows)

	summaries, err := repo.GetLogsByCaregiver(context.Background(), "caregiver-1", repository.CaregiverLogFilter{})
	if err != nil {
		t.Fatalf("GetLogsByCaregiver error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ID != "log-1" || summaries[0].LogType != domain.LogTypeClockIn {
		t.Fatalf("unexpected first summary: %+v", summaries[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryGetLogsByCaregiverWithFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	logType := domain.LogTypeClockIn
	startDate := now.Add(-24 * time.Hour)
	endDate := now
	filter := repository.CaregiverLogFilter{
		LogType:   &logType,
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     10,
		Offset:    5,
	}

	rows := sqlmock.NewRows([]string{"id", "caregiver_id", "log_type", "timestamp"}).
		AddRow("log-1", "caregiver-1", "clock_in", now)

	mock.ExpectQuery("SELECT id, caregiver_id, log_type, timestamp FROM logs_caregivers WHERE caregiver_id = \\$1 AND log_type = \\$2 AND timestamp >= \\$3 AND timestamp <= \\$4 ORDER BY timestamp DESC LIMIT \\$5 OFFSET \\$6").
		WithArgs("caregiver-1", *filter.LogType, *filter.StartDate, *filter.EndDate, filter.Limit, filter.Offset).
		WillReturnRows(rows)

	summaries, err := repo.GetLogsByCaregiver(context.Background(), "caregiver-1", filter)
	if err != nil {
		t.Fatalf("GetLogsByCaregiver error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryGetTodayLogs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	_ = today.Add(24 * time.Hour) // tomorrow calculation for reference

	rows := sqlmock.NewRows([]string{"id", "caregiver_id", "log_type", "latitude", "longitude", "timestamp", "notes", "created_at"}).
		AddRow("log-1", "caregiver-1", "clock_in", 40.7128, -74.0060, today.Add(9*time.Hour), nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
        FROM logs_caregivers
        WHERE caregiver_id = $1 AND timestamp >= $2 AND timestamp < $3
        ORDER BY timestamp ASC
    `)).
		WithArgs("caregiver-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	logs, err := repo.GetTodayLogs(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("GetTodayLogs error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].ID != "log-1" || logs[0].LogType != domain.LogTypeClockIn {
		t.Fatalf("unexpected log: %+v", logs[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryGetLastClockIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "caregiver_id", "log_type", "latitude", "longitude", "timestamp", "notes", "created_at"}).
		AddRow("log-1", "caregiver-1", "clock_in", 40.7128, -74.0060, now, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
        FROM logs_caregivers
        WHERE caregiver_id = $1 AND log_type = 'clock_in'
        ORDER BY timestamp DESC
        LIMIT 1
    `)).
		WithArgs("caregiver-1").
		WillReturnRows(rows)

	log, err := repo.GetLastClockIn(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("GetLastClockIn error: %v", err)
	}
	if log == nil || log.ID != "log-1" || log.LogType != domain.LogTypeClockIn {
		t.Fatalf("unexpected log: %+v", log)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryGetLastClockInNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
        FROM logs_caregivers
        WHERE caregiver_id = $1 AND log_type = 'clock_in'
        ORDER BY timestamp DESC
        LIMIT 1
    `)).
		WithArgs("caregiver-1").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetLastClockIn(context.Background(), "caregiver-1")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryGetLastClockOut(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "caregiver_id", "log_type", "latitude", "longitude", "timestamp", "notes", "created_at"}).
		AddRow("log-2", "caregiver-1", "clock_out", 40.7128, -74.0060, now, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
        FROM logs_caregivers
        WHERE caregiver_id = $1 AND log_type = 'clock_out'
        ORDER BY timestamp DESC
        LIMIT 1
    `)).
		WithArgs("caregiver-1").
		WillReturnRows(rows)

	log, err := repo.GetLastClockOut(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("GetLastClockOut error: %v", err)
	}
	if log == nil || log.ID != "log-2" || log.LogType != domain.LogTypeClockOut {
		t.Fatalf("unexpected log: %+v", log)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryHasClockedInToday(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT EXISTS(
            SELECT 1 FROM logs_caregivers
            WHERE caregiver_id = $1 AND log_type = 'clock_in'
            AND timestamp >= $2 AND timestamp < $3
        )
    `)).
		WithArgs("caregiver-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	hasClockedIn, err := repo.HasClockedInToday(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("HasClockedInToday error: %v", err)
	}
	if !hasClockedIn {
		t.Fatalf("expected hasClockedIn to be true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverLogRepositoryHasClockedOutToday(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverLogRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT EXISTS(
            SELECT 1 FROM logs_caregivers
            WHERE caregiver_id = $1 AND log_type = 'clock_out'
            AND timestamp >= $2 AND timestamp < $3
        )
    `)).
		WithArgs("caregiver-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	hasClockedOut, err := repo.HasClockedOutToday(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("HasClockedOutToday error: %v", err)
	}
	if hasClockedOut {
		t.Fatalf("expected hasClockedOut to be false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}