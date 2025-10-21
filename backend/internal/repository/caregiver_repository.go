package repository

import (
	"context"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
)

// CaregiverRepository exposes read operations for caregivers.
type CaregiverRepository interface {
	GetByID(ctx context.Context, caregiverID string) (domain.Caregiver, error)
}
