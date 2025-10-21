package domain

import "time"

// AuthClient models an OAuth2/OIDC style client credential.
type AuthClient struct {
	ID          string
	SecretHash  string
	Description string
	CaregiverID string
	Scopes      []string
}

// TokenPair holds the generated access and id tokens issued to a client.
type TokenPair struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	Scope       string    `json:"scope,omitempty"`
	IDToken     string    `json:"id_token,omitempty"`
	IssuedAt    time.Time `json:"-"`
}
