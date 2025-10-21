package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

func TestScheduleRepositoryListSchedules(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewScheduleRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "caregiver_id", "client_name", "service_name", "start_time", "end_time", "status", "location_label"}).
		AddRow("sched-1", "cg-1", "Client A", "Service", time.Now(), time.Now().Add(time.Hour), "scheduled", "Location")

	mock.ExpectQuery("SELECT s\\.id[\\s\\S]+ORDER BY s\\.start_time ASC").
		WithArgs("cg-1").
		WillReturnRows(rows)

	summaries, err := repo.ListSchedules(context.Background(), "cg-1", repository.ScheduleFilter{})
	if err != nil {
		t.Fatalf("ListSchedules error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ID != "sched-1" {
		t.Fatalf("unexpected schedule id: %s", summaries[0].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestScheduleRepositoryGetScheduleForCaregiver(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewScheduleRepository(sqlxDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "caregiver_id", "client_id", "service_name", "location_label", "start_time", "end_time", "status",
		"clock_in_at", "clock_in_lat", "clock_in_long", "clock_out_at", "clock_out_lat", "clock_out_long", "notes",
		"created_at", "updated_at",
		"client_full_name", "client_email", "client_phone", "client_address", "client_city", "client_state", "client_postal", "client_latitude", "client_longitude",
	}).AddRow(
		"sched-1", "cg-1", "client-1", "Service", "Location", now, now.Add(time.Hour), "scheduled",
		nil, nil, nil, nil, nil, nil, nil,
		now, now,
		"Client A", "client@example.com", "123", "Addr", "City", "State", "Postal", 1.23, 4.56,
	)

	mock.ExpectQuery("SELECT s\\.\\*[\\s\\S]+WHERE s\\.id = \\$1 AND s\\.caregiver_id = \\$2").
		WithArgs("sched-1", "cg-1").
		WillReturnRows(rows)

	schedule, err := repo.GetScheduleForCaregiver(context.Background(), "sched-1", "cg-1")
	if err != nil {
		t.Fatalf("GetScheduleForCaregiver error: %v", err)
	}
	if schedule.ID != "sched-1" || schedule.Client.FullName != "Client A" {
		t.Fatalf("unexpected schedule result: %+v", schedule)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestScheduleRepositoryLogClockIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewScheduleRepository(sqlxDB)

	now := time.Now()
	event := domain.VisitEvent{Timestamp: now, Latitude: 1.23, Longitude: 4.56}

	mock.ExpectExec(regexp.QuoteMeta(`
        UPDATE schedules
        SET clock_in_at = $2,
            clock_in_lat = $3,
            clock_in_long = $4,
            updated_at = NOW()
        WHERE id = $1
    `)).
		WithArgs("sched-1", event.Timestamp, event.Latitude, event.Longitude).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.LogClockIn(context.Background(), "sched-1", event); err != nil {
		t.Fatalf("LogClockIn error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
