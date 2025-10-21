package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

func TestCaregiverRepositoryGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow("caregiver-1", "John Doe", "john@example.com")

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, name, email
        FROM caregivers
        WHERE id = $1
    `)).
		WithArgs("caregiver-1").
		WillReturnRows(rows)

	caregiver, err := repo.GetByID(context.Background(), "caregiver-1")
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if caregiver.ID != "caregiver-1" || caregiver.Name != "John Doe" || caregiver.Email != "john@example.com" {
		t.Fatalf("unexpected caregiver returned: %+v", caregiver)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverRepositoryGetByIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, name, email
        FROM caregivers
        WHERE id = $1
    `)).
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), "non-existent")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCaregiverRepositoryGetByIDDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewCaregiverRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, name, email
        FROM caregivers
        WHERE id = $1
    `)).
		WithArgs("caregiver-1").
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetByID(context.Background(), "caregiver-1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}