package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ScheduleRepository implements repository.ScheduleRepository backed by Postgres.
type ScheduleRepository struct {
	db *sqlx.DB
}

// NewScheduleRepository creates a new repository.
func NewScheduleRepository(db *sqlx.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

type scheduleSummaryRow struct {
	ID           string    `db:"id"`
	CaregiverID  string    `db:"caregiver_id"`
	ClientName   string    `db:"client_name"`
	ServiceName  string    `db:"service_name"`
	StartTime    time.Time `db:"start_time"`
	EndTime      time.Time `db:"end_time"`
	Status       string    `db:"status"`
	LocationName string    `db:"location_label"`
}

func (r *ScheduleRepository) ListSchedules(ctx context.Context, caregiverID string, filter repository.ScheduleFilter) ([]domain.ScheduleSummary, error) {
	query := `
		SELECT s.id,
		       s.caregiver_id,
		       c.full_name AS client_name,
		       s.service_name,
		       s.start_time,
		       s.end_time,
		       s.status,
		       s.location_label
		FROM schedules s
		INNER JOIN clients c ON c.id = s.client_id
		WHERE s.caregiver_id = $1
	`
	args := []interface{}{caregiverID}
	argPosition := 2

	if len(filter.Status) > 0 {
		query += " AND s.status = ANY($" + itoa(argPosition) + ")"
		argPosition++
		statusStrings := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			statusStrings[i] = string(status)
		}
		args = append(args, pq.Array(statusStrings))
	}

	if filter.Date != nil {
		start := filter.Date.Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)
		query += " AND s.start_time >= $" + itoa(argPosition)
		args = append(args, start)
		argPosition++
		query += " AND s.start_time < $" + itoa(argPosition)
		args = append(args, end)
		argPosition++
	}

	query += ` ORDER BY 
		CASE s.status 
			WHEN 'in_progress' THEN 0
			WHEN 'scheduled' THEN 1
			WHEN 'missed' THEN 2
			WHEN 'completed' THEN 3
			WHEN 'cancelled' THEN 4
			ELSE 5
		END,
		s.start_time DESC`

	// Add pagination if limit is specified
	if filter.Limit > 0 {
		query += " LIMIT $" + itoa(argPosition)
		args = append(args, filter.Limit)
		argPosition++

		if filter.Offset > 0 {
			query += " OFFSET $" + itoa(argPosition)
			args = append(args, filter.Offset)
		}
	}

	rows := []scheduleSummaryRow{}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make([]domain.ScheduleSummary, len(rows))
	for i, row := range rows {
		result[i] = domain.ScheduleSummary{
			ID:           row.ID,
			CaregiverID:  row.CaregiverID,
			ClientName:   row.ClientName,
			ServiceName:  row.ServiceName,
			StartTime:    row.StartTime,
			EndTime:      row.EndTime,
			Status:       domain.ScheduleStatus(row.Status),
			LocationName: row.LocationName,
		}
	}
	return result, nil
}

type scheduleRow struct {
	ID           string          `db:"id"`
	CaregiverID  string          `db:"caregiver_id"`
	ClientID     string          `db:"client_id"`
	ServiceName  string          `db:"service_name"`
	LocationName sql.NullString  `db:"location_label"`
	StartTime    time.Time       `db:"start_time"`
	EndTime      time.Time       `db:"end_time"`
	Status       string          `db:"status"`
	ClockInAt    sql.NullTime    `db:"clock_in_at"`
	ClockInLat   sql.NullFloat64 `db:"clock_in_lat"`
	ClockInLong  sql.NullFloat64 `db:"clock_in_long"`
	ClockOutAt   sql.NullTime    `db:"clock_out_at"`
	ClockOutLat  sql.NullFloat64 `db:"clock_out_lat"`
	ClockOutLong sql.NullFloat64 `db:"clock_out_long"`
	Notes        sql.NullString  `db:"notes"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`

	ClientFullName string          `db:"client_full_name"`
	ClientEmail    sql.NullString  `db:"client_email"`
	ClientPhone    sql.NullString  `db:"client_phone"`
	ClientAddress  sql.NullString  `db:"client_address"`
	ClientCity     sql.NullString  `db:"client_city"`
	ClientState    sql.NullString  `db:"client_state"`
	ClientPostal   sql.NullString  `db:"client_postal"`
	ClientLat      sql.NullFloat64 `db:"client_latitude"`
	ClientLong     sql.NullFloat64 `db:"client_longitude"`
}

func (r *ScheduleRepository) GetSchedule(ctx context.Context, scheduleID string) (domain.Schedule, error) {
	return r.fetchSchedule(ctx, "s.id = $1", scheduleID)
}

func (r *ScheduleRepository) GetScheduleForCaregiver(ctx context.Context, scheduleID, caregiverID string) (domain.Schedule, error) {
	return r.fetchSchedule(ctx, "s.id = $1 AND s.caregiver_id = $2", scheduleID, caregiverID)
}

func (r *ScheduleRepository) fetchSchedule(ctx context.Context, predicate string, args ...interface{}) (domain.Schedule, error) {
	query := `
		SELECT s.*,
		       c.full_name AS client_full_name,
		       c.email AS client_email,
		       c.phone AS client_phone,
		       c.address AS client_address,
		       c.city AS client_city,
		       c.state AS client_state,
		       c.postal_code AS client_postal,
		       c.latitude AS client_latitude,
		       c.longitude AS client_longitude
		FROM schedules s
		INNER JOIN clients c ON c.id = s.client_id
		WHERE ` + predicate

	var row scheduleRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Schedule{}, domain.ErrNotFound
		}
		return domain.Schedule{}, err
	}

	return mapSchedule(row), nil
}

func (r *ScheduleRepository) LogClockIn(ctx context.Context, scheduleID string, event domain.VisitEvent) error {
	query := `
		UPDATE schedules
		SET clock_in_at = $2,
		    clock_in_lat = $3,
		    clock_in_long = $4,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, scheduleID, event.Timestamp, event.Latitude, event.Longitude)
	return err
}

func (r *ScheduleRepository) LogClockOut(ctx context.Context, scheduleID string, event domain.VisitEvent) error {
	query := `
		UPDATE schedules
		SET clock_out_at = $2,
		    clock_out_lat = $3,
		    clock_out_long = $4,
		    notes = COALESCE($5, notes),
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, scheduleID, event.Timestamp, event.Latitude, event.Longitude, event.Notes)
	return err
}

func (r *ScheduleRepository) UpdateStatus(ctx context.Context, scheduleID string, status domain.ScheduleStatus) error {
	query := `
		UPDATE schedules
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, scheduleID, status)
	return err
}

func (r *ScheduleRepository) GetMetrics(ctx context.Context, caregiverID string, day time.Time) (domain.ScheduleMetrics, error) {
	start := day.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)

	type counts struct {
		Total      int `db:"total"`
		Scheduled  int `db:"scheduled"`
		InProgress int `db:"in_progress"`
		Completed  int `db:"completed"`
		Cancelled  int `db:"cancelled"`
		Missed     int `db:"missed"`
	}

	query := `
		SELECT
			COUNT(*) FILTER (WHERE s.start_time >= $2 AND s.start_time < $3) AS total,
			COUNT(*) FILTER (WHERE s.status = 'scheduled' AND s.start_time >= $2 AND s.start_time < $3) AS scheduled,
			COUNT(*) FILTER (WHERE s.status = 'in_progress' AND s.start_time >= $2 AND s.start_time < $3) AS in_progress,
			COUNT(*) FILTER (WHERE s.status = 'completed' AND s.start_time >= $2 AND s.start_time < $3) AS completed,
			COUNT(*) FILTER (WHERE s.status = 'cancelled' AND s.start_time >= $2 AND s.start_time < $3) AS cancelled,
			COUNT(*) FILTER (WHERE s.status = 'missed' AND s.start_time >= $2 AND s.start_time < $3) AS missed
		FROM schedules s
		WHERE s.caregiver_id = $1
	`

	var c counts
	if err := r.db.GetContext(ctx, &c, query, caregiverID, start, end); err != nil {
		return domain.ScheduleMetrics{}, err
	}

	return domain.ScheduleMetrics{
		Total:      c.Total,
		Upcoming:   c.Scheduled,
		InProgress: c.InProgress,
		Completed:  c.Completed,
		Cancelled:  c.Cancelled,
		Missed:     c.Missed,
	}, nil
}

func mapSchedule(row scheduleRow) domain.Schedule {
	var clockInAt, clockOutAt *time.Time
	if row.ClockInAt.Valid {
		t := row.ClockInAt.Time
		clockInAt = &t
	}
	if row.ClockOutAt.Valid {
		t := row.ClockOutAt.Time
		clockOutAt = &t
	}

	var clockInLat, clockInLong, clockOutLat, clockOutLong *float64
	if row.ClockInLat.Valid {
		val := row.ClockInLat.Float64
		clockInLat = &val
	}
	if row.ClockInLong.Valid {
		val := row.ClockInLong.Float64
		clockInLong = &val
	}
	if row.ClockOutLat.Valid {
		val := row.ClockOutLat.Float64
		clockOutLat = &val
	}
	if row.ClockOutLong.Valid {
		val := row.ClockOutLong.Float64
		clockOutLong = &val
	}

	var notes *string
	if row.Notes.Valid {
		val := row.Notes.String
		notes = &val
	}

	duration := int(row.EndTime.Sub(row.StartTime).Minutes())

	client := domain.Client{
		ID:       row.ClientID,
		FullName: row.ClientFullName,
	}
	if row.ClientEmail.Valid {
		client.Email = row.ClientEmail.String
	}
	if row.ClientPhone.Valid {
		client.Phone = row.ClientPhone.String
	}
	if row.ClientAddress.Valid {
		client.Address = row.ClientAddress.String
	}
	if row.ClientCity.Valid {
		client.City = row.ClientCity.String
	}
	if row.ClientState.Valid {
		client.State = row.ClientState.String
	}
	if row.ClientPostal.Valid {
		client.Postal = row.ClientPostal.String
	}
	if row.ClientLat.Valid {
		client.Latitude = row.ClientLat.Float64
	}
	if row.ClientLong.Valid {
		client.Longitude = row.ClientLong.Float64
	}

	var location string
	if row.LocationName.Valid {
		location = row.LocationName.String
	}

	return domain.Schedule{
		ID:            row.ID,
		CaregiverID:   row.CaregiverID,
		Client:        client,
		ServiceName:   row.ServiceName,
		LocationLabel: location,
		StartTime:     row.StartTime,
		EndTime:       row.EndTime,
		Status:        domain.ScheduleStatus(row.Status),
		ClockInAt:     clockInAt,
		ClockInLat:    clockInLat,
		ClockInLong:   clockInLong,
		ClockOutAt:    clockOutAt,
		ClockOutLat:   clockOutLat,
		ClockOutLong:  clockOutLong,
		Notes:         notes,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		DurationMins:  duration,
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
