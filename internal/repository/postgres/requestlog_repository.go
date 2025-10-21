package postgres

import (
	"context"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

// RequestLogRepository implements repository.RequestLogRepository using Postgres.
type RequestLogRepository struct {
	db *sqlx.DB
}

// NewRequestLogRepository constructs a repository.
func NewRequestLogRepository(db *sqlx.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

// Insert persists the request log.
func (r *RequestLogRepository) Insert(ctx context.Context, log domain.RequestLog) error {
	query := `
		INSERT INTO request_logs (method, path, query, status, latency_ms, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.Method,
		log.Path,
		log.Query,
		log.Status,
		log.Latency.Milliseconds(),
		log.IP,
		log.UserAgent,
	)
	return err
}
