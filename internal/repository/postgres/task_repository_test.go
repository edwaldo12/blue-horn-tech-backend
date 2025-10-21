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

func TestTaskRepositoryListBySchedule(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewTaskRepository(sqlxDB)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "schedule_id", "title", "description", "status", "not_completed_reason", "sort_order", "updated_at",
	}).AddRow("task-1", "sched-1", "Check vital", "", "pending", nil, 1, now)

	mock.ExpectQuery("SELECT id[\\s\\S]+FROM schedule_tasks[\\s\\S]+WHERE schedule_id = \\$1[\\s\\S]+ORDER BY sort_order ASC, created_at ASC").
		WithArgs("sched-1").
		WillReturnRows(rows)

	tasks, err := repo.ListBySchedule(context.Background(), "sched-1")
	if err != nil {
		t.Fatalf("ListBySchedule error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "task-1" {
		t.Fatalf("unexpected tasks result: %+v", tasks)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskRepositoryUpdateTaskStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewTaskRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(`
        UPDATE schedule_tasks
        SET status = $2,
            not_completed_reason = $3,
            updated_at = NOW()
        WHERE id = $1
    `)).
		WithArgs("task-1", domain.TaskStatusCompleted, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateTaskStatus(context.Background(), "task-1", domain.TaskStatusCompleted, nil); err != nil {
		t.Fatalf("UpdateTaskStatus error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskRepositoryCreateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "pgx")
	repo := NewTaskRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO schedule_tasks (id, schedule_id, title, description, status, not_completed_reason, sort_order)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, COALESCE($6, 0))
        RETURNING id
    `)).
		WithArgs("sched-1", "New Task", nil, domain.TaskStatusPending, nil, int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("task-123"))

	id, err := repo.CreateTask(context.Background(), domain.Task{ScheduleID: "sched-1", Title: "New Task", Status: domain.TaskStatusPending})
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected returned id")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
