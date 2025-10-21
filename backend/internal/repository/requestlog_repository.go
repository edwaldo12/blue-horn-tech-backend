package repository

import (
	"context"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
)

// RequestLogRepository persists request log entries.
type RequestLogRepository interface {
	Insert(ctx context.Context, log domain.RequestLog) error
}
