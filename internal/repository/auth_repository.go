package repository

import (
	"context"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
)

// AuthRepository encapsulates persistence for OAuth2/OIDC style clients.
type AuthRepository interface {
	//
	GetClientByID(ctx context.Context, clientID string) (domain.AuthClient, error)
}
