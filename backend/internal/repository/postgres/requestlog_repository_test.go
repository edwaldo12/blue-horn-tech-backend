package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

func TestRequestLogRepositoryInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewRequestLogRepository(sqlxDB)

	entry := domain.RequestLog{Method: "GET", Path: "/api", Query: "", Status: 200, Latency: 50 * time.Millisecond, IP: "127.0.0.1", UserAgent: "test"}

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO request_logs (method, path, query, status, latency_ms, ip_address, user_agent)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `)).
		WithArgs(entry.Method, entry.Path, entry.Query, entry.Status, entry.Latency.Milliseconds(), entry.IP, entry.UserAgent).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Insert(context.Background(), entry); err != nil {
		t.Fatalf("Insert error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
