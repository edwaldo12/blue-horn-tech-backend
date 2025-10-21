package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// AuthRepository implements repository.AuthRepository.
type AuthRepository struct {
	db *sqlx.DB
}

// NewAuthRepository builds the repository.
func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

type authClientRow struct {
	ID          string         `db:"id"`
	SecretHash  string         `db:"secret_hash"`
	Description sql.NullString `db:"description"`
	CaregiverID string         `db:"caregiver_id"`
	Scopes      pq.StringArray `db:"scopes"`
}

func (r *AuthRepository) GetClientByID(ctx context.Context, clientID string) (domain.AuthClient, error) {
	query := `
		SELECT id,
		       secret_hash,
		       description,
		       caregiver_id,
		       scopes
		FROM auth_clients
		WHERE id = $1
	`

	var row authClientRow
	if err := r.db.GetContext(ctx, &row, query, clientID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AuthClient{}, domain.ErrNotFound
		}
		return domain.AuthClient{}, err
	}

	client := domain.AuthClient{
		ID:          row.ID,
		SecretHash:  row.SecretHash,
		CaregiverID: row.CaregiverID,
		Scopes:      []string(row.Scopes),
	}
	if row.Description.Valid {
		client.Description = row.Description.String
	}
	return client, nil
}
