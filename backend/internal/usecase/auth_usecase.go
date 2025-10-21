package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/config"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TokenRequest models the OAuth2 style request payload.
type TokenRequest struct {
	GrantType    string
	ClientID     string
	ClientSecret string
	Scope        string
}

// AuthUsecase encapsulates token issuing and verification logic.
type AuthUsecase struct {
	clients    repository.AuthRepository
	caregivers repository.CaregiverRepository
	cfg        config.AuthConfig
	now        func() time.Time
}

// NewAuthUsecase constructs the usecase.
func NewAuthUsecase(
	cfg config.AuthConfig,
	clients repository.AuthRepository,
	caregivers repository.CaregiverRepository,
) *AuthUsecase {
	return &AuthUsecase{
		clients:    clients,
		caregivers: caregivers,
		cfg:        cfg,
		now:        time.Now,
	}
}

// IssueToken handles a simplified client credentials flow returning JWT tokens.
func (uc *AuthUsecase) IssueToken(ctx context.Context, req TokenRequest) (domain.TokenPair, domain.Caregiver, error) {
	if strings.ToLower(req.GrantType) != "client_credentials" {
		return domain.TokenPair{}, domain.Caregiver{}, domain.ErrValidationFailure
	}
	client, err := uc.clients.GetClientByID(ctx, req.ClientID)
	if err != nil {
		return domain.TokenPair{}, domain.Caregiver{}, err
	}

	if !uc.verifySecret(client.SecretHash, req.ClientSecret) {
		return domain.TokenPair{}, domain.Caregiver{}, domain.ErrUnauthorized
	}

	caregiver, err := uc.caregivers.GetByID(ctx, client.CaregiverID)
	if err != nil {
		return domain.TokenPair{}, domain.Caregiver{}, err
	}

	scope := req.Scope
	if scope == "" {
		scope = strings.Join(client.Scopes, " ")
	}

	issuedAt := uc.now()
	accessToken, err := uc.buildAccessToken(client, caregiver, scope, issuedAt)
	if err != nil {
		return domain.TokenPair{}, domain.Caregiver{}, err
	}

	idToken, err := uc.buildIDToken(client, caregiver, scope, issuedAt)
	if err != nil {
		return domain.TokenPair{}, domain.Caregiver{}, err
	}

	pair := domain.TokenPair{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(uc.cfg.AccessTokenTTL.Seconds()),
		Scope:       scope,
		IDToken:     idToken,
		IssuedAt:    issuedAt,
	}

	return pair, caregiver, nil
}

func (uc *AuthUsecase) verifySecret(hash, secret string) bool {
	if strings.HasPrefix(hash, "bcrypt$") {
		encoded := strings.TrimPrefix(hash, "bcrypt$")
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return false
		}
		if err := bcrypt.CompareHashAndPassword(data, []byte(secret)); err != nil {
			return false
		}
		return true
	}

	h := sha256.Sum256([]byte(secret))
	return subtleConstantTimeCompare(hash, fmt.Sprintf("%x", h[:]))
}

func (uc *AuthUsecase) buildAccessToken(client domain.AuthClient, caregiver domain.Caregiver, scope string, issuedAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss":   uc.cfg.Issuer,
		"sub":   caregiver.ID,
		"aud":   uc.cfg.Audience,
		"iat":   issuedAt.Unix(),
		"exp":   issuedAt.Add(uc.cfg.AccessTokenTTL).Unix(),
		"scope": scope,
		"cid":   client.ID,
		"cname": caregiver.Name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.cfg.AccessTokenSecret))
}

func (uc *AuthUsecase) buildIDToken(client domain.AuthClient, caregiver domain.Caregiver, scope string, issuedAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss":   uc.cfg.Issuer,
		"sub":   caregiver.ID,
		"aud":   []string{uc.cfg.Audience},
		"iat":   issuedAt.Unix(),
		"exp":   issuedAt.Add(uc.cfg.IDTokenTTL).Unix(),
		"name":  caregiver.Name,
		"email": caregiver.Email,
		"scope": scope,
		"cid":   client.ID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.cfg.IDTokenSecret))
}

// ParseToken validates the incoming bearer token and returns the caregiver ID.
func (uc *AuthUsecase) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	if tokenStr == "" {
		return nil, domain.ErrUnauthorized
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(uc.cfg.AccessTokenSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func subtleConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
