package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

// CaregiverRepository implements repository.CaregiverRepository.
type CaregiverRepository struct {
	db *sqlx.DB
}

// NewCaregiverRepository constructs the repository.
func NewCaregiverRepository(db *sqlx.DB) *CaregiverRepository {
	return &CaregiverRepository{db: db}
}

func (r *CaregiverRepository) GetByID(ctx context.Context, caregiverID string) (domain.Caregiver, error) {
	query := `
		SELECT id, name, email
		FROM caregivers
		WHERE id = $1
	`
	var caregiver domain.Caregiver
	if err := r.db.GetContext(ctx, &caregiver, query, caregiverID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Caregiver{}, domain.ErrNotFound
		}
		return domain.Caregiver{}, err
	}
	return caregiver, nil
}
