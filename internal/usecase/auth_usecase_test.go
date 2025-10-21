package usecase

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/config"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type authRepoStub struct {
	client domain.AuthClient
}

func (a *authRepoStub) GetClientByID(ctx context.Context, clientID string) (domain.AuthClient, error) {
	if a.client.ID != clientID {
		return domain.AuthClient{}, domain.ErrNotFound
	}
	return a.client, nil
}

type caregiverRepoStub struct {
	caregiver domain.Caregiver
}

func (c *caregiverRepoStub) GetByID(ctx context.Context, caregiverID string) (domain.Caregiver, error) {
	if c.caregiver.ID != caregiverID {
		return domain.Caregiver{}, domain.ErrNotFound
	}
	return c.caregiver, nil
}

func TestAuthUsecaseIssueToken(t *testing.T) {
	hashBytes, err := bcryptGenerate("secret")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	client := domain.AuthClient{
		ID:          "client-1",
		SecretHash:  hashBytes,
		CaregiverID: "care-1",
		Scopes:      []string{"schedules.read"},
	}
	authRepo := &authRepoStub{client: client}
	careRepo := &caregiverRepoStub{caregiver: domain.Caregiver{ID: "care-1", Name: "Louis", Email: "louis@example.com"}}

	cfg := config.AuthConfig{
		Issuer:            "http://localhost:8080",
		Audience:          "caregiver-app",
		AccessTokenSecret: "access-secret",
		IDTokenSecret:     "id-secret",
		AccessTokenTTL:    time.Minute,
		IDTokenTTL:        time.Hour,
	}

	uc := NewAuthUsecase(cfg, authRepo, careRepo)

	pair, caregiver, err := uc.IssueToken(context.Background(), TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client-1",
		ClientSecret: "secret",
	})
	if err != nil {
		t.Fatalf("issue token error: %v", err)
	}

	if caregiver.ID != "care-1" {
		t.Fatalf("unexpected caregiver returned: %v", caregiver.ID)
	}

	claims, err := uc.ParseToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse token error: %v", err)
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub != "care-1" {
		t.Fatalf("expected sub claim care-1, got %v", claims["sub"])
	}

	if pair.Scope == "" {
		t.Fatalf("expected default scope populated")
	}
}

func bcryptGenerate(secret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return "bcrypt$" + base64.StdEncoding.EncodeToString(hash), nil
}
