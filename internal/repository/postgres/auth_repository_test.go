package postgres

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func TestAuthRepositoryGetClientByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewAuthRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "secret_hash", "description", "caregiver_id", "scopes"}).
		AddRow("client-1", "hash", "desc", "caregiver", pq.StringArray{"scope1"})

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id,
               secret_hash,
               description,
               caregiver_id,
               scopes
        FROM auth_clients
        WHERE id = $1
    `)).
		WithArgs("client-1").
		WillReturnRows(rows)

	client, err := repo.GetClientByID(context.Background(), "client-1")
	if err != nil {
		t.Fatalf("GetClientByID error: %v", err)
	}
	if client.ID != "client-1" || client.CaregiverID != "caregiver" {
		t.Fatalf("unexpected client returned: %+v", client)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
