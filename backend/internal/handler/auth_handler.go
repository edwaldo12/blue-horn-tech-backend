package handler

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes OAuth style endpoints.
type AuthHandler struct {
	authUC *usecase.AuthUsecase
}

// NewAuthHandler creates a new handler.
func NewAuthHandler(authUC *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

type tokenRequestJSON struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
}

// Token issues an access token following a simplified client credentials flow.
func (h *AuthHandler) Token(c *gin.Context) {
	req := tokenRequestJSON{
		GrantType: strings.TrimSpace(c.PostForm("grant_type")),
		ClientID:  strings.TrimSpace(c.PostForm("client_id")),
		Scope:     strings.TrimSpace(c.PostForm("scope")),
	}
	clientSecret := strings.TrimSpace(c.PostForm("client_secret"))

	// If JSON body is supplied, bind it.
	if c.GetHeader("Content-Type") == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrValidationFailure.Error(), "message": err.Error()})
			return
		}
		clientSecret = req.ClientSecret
	}

	// Support Basic auth header.
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
			raw := strings.TrimSpace(authHeader[len("Basic "):])
			decoded, err := base64.StdEncoding.DecodeString(raw)
			if err == nil {
				parts := strings.SplitN(string(decoded), ":", 2)
				if len(parts) == 2 {
					req.ClientID = parts[0]
					clientSecret = parts[1]
				}
			}
		}
	}

	if req.GrantType == "" {
		req.GrantType = "client_credentials"
	}

	tokenPair, caregiver, err := h.authUC.IssueToken(c, usecase.TokenRequest{
		GrantType:    req.GrantType,
		ClientID:     req.ClientID,
		ClientSecret: clientSecret,
		Scope:        req.Scope,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenPair.AccessToken,
		"token_type":   tokenPair.TokenType,
		"expires_in":   tokenPair.ExpiresIn,
		"scope":        tokenPair.Scope,
		"id_token":     tokenPair.IDToken,
		"caregiver": gin.H{
			"id":    caregiver.ID,
			"name":  caregiver.Name,
			"email": caregiver.Email,
		},
	})
}
