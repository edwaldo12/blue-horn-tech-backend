package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/jmoiron/sqlx"
)


// CaregiverLogRepository implements repository.CaregiverLogRepository backed by Postgres.
type CaregiverLogRepository struct {
	db *sqlx.DB
}

// NewCaregiverLogRepository creates a new caregiver log repository.
func NewCaregiverLogRepository(db *sqlx.DB) *CaregiverLogRepository {
	return &CaregiverLogRepository{db: db}
}

type caregiverLogRow struct {
	ID          string         `db:"id"`
	CaregiverID string         `db:"caregiver_id"`
	LogType     string         `db:"log_type"`
	Latitude    float64        `db:"latitude"`
	Longitude   float64        `db:"longitude"`
	Timestamp   time.Time      `db:"timestamp"`
	Notes       sql.NullString `db:"notes"`
	CreatedAt   time.Time      `db:"created_at"`
}

func (r *CaregiverLogRepository) CreateLog(ctx context.Context, log domain.CaregiverLog) (string, error) {
	query := `
		INSERT INTO logs_caregivers (caregiver_id, log_type, latitude, longitude, timestamp, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		log.CaregiverID,
		log.LogType,
		log.Latitude,
		log.Longitude,
		log.Timestamp,
		log.Notes,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *CaregiverLogRepository) GetLogsByCaregiver(ctx context.Context, caregiverID string, filter repository.CaregiverLogFilter) ([]domain.CaregiverLogSummary, error) {
	query := `
		SELECT id, caregiver_id, log_type, timestamp
		FROM logs_caregivers
		WHERE caregiver_id = $1
	`
	args := []interface{}{caregiverID}
	argPosition := 2

	if filter.LogType != nil {
		query += " AND log_type = $" + itoa(argPosition)
		args = append(args, *filter.LogType)
		argPosition++
	}

	if filter.StartDate != nil {
		query += " AND timestamp >= $" + itoa(argPosition)
		args = append(args, *filter.StartDate)
		argPosition++
	}

	if filter.EndDate != nil {
		query += " AND timestamp <= $" + itoa(argPosition)
		args = append(args, *filter.EndDate)
		argPosition++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT $" + itoa(argPosition)
		args = append(args, filter.Limit)
		argPosition++

		if filter.Offset > 0 {
			query += " OFFSET $" + itoa(argPosition)
			args = append(args, filter.Offset)
		}
	}

	rows := []caregiverLogRow{}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make([]domain.CaregiverLogSummary, len(rows))
	for i, row := range rows {
		result[i] = domain.CaregiverLogSummary{
			ID:          row.ID,
			CaregiverID: row.CaregiverID,
			LogType:     domain.LogType(row.LogType),
			Timestamp:   row.Timestamp,
		}
	}

	return result, nil
}

func (r *CaregiverLogRepository) GetTodayLogs(ctx context.Context, caregiverID string) ([]domain.CaregiverLog, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	query := `
		SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
		FROM logs_caregivers
		WHERE caregiver_id = $1 AND timestamp >= $2 AND timestamp < $3
		ORDER BY timestamp ASC
	`

	rows := []caregiverLogRow{}
	err := r.db.SelectContext(ctx, &rows, query, caregiverID, today, tomorrow)
	if err != nil {
		return nil, err
	}

	result := make([]domain.CaregiverLog, len(rows))
	for i, row := range rows {
		result[i] = mapCaregiverLog(row)
	}

	return result, nil
}

func (r *CaregiverLogRepository) GetLastClockIn(ctx context.Context, caregiverID string) (*domain.CaregiverLog, error) {
	query := `
		SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
		FROM logs_caregivers
		WHERE caregiver_id = $1 AND log_type = 'clock_in'
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var row caregiverLogRow
	err := r.db.GetContext(ctx, &row, query, caregiverID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	log := mapCaregiverLog(row)
	return &log, nil
}

func (r *CaregiverLogRepository) GetLastClockOut(ctx context.Context, caregiverID string) (*domain.CaregiverLog, error) {
	query := `
		SELECT id, caregiver_id, log_type, latitude, longitude, timestamp, notes, created_at
		FROM logs_caregivers
		WHERE caregiver_id = $1 AND log_type = 'clock_out'
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var row caregiverLogRow
	err := r.db.GetContext(ctx, &row, query, caregiverID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	log := mapCaregiverLog(row)
	return &log, nil
}

func (r *CaregiverLogRepository) HasClockedInToday(ctx context.Context, caregiverID string) (bool, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM logs_caregivers
			WHERE caregiver_id = $1 AND log_type = 'clock_in'
			AND timestamp >= $2 AND timestamp < $3
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, caregiverID, today, tomorrow)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *CaregiverLogRepository) HasClockedOutToday(ctx context.Context, caregiverID string) (bool, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM logs_caregivers
			WHERE caregiver_id = $1 AND log_type = 'clock_out'
			AND timestamp >= $2 AND timestamp < $3
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, caregiverID, today, tomorrow)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func mapCaregiverLog(row caregiverLogRow) domain.CaregiverLog {
	var notes *string
	if row.Notes.Valid {
		val := row.Notes.String
		notes = &val
	}

	return domain.CaregiverLog{
		ID:          row.ID,
		CaregiverID: row.CaregiverID,
		LogType:     domain.LogType(row.LogType),
		Latitude:    row.Latitude,
		Longitude:   row.Longitude,
		Timestamp:   row.Timestamp,
		Notes:       notes,
		CreatedAt:   row.CreatedAt,
	}
}