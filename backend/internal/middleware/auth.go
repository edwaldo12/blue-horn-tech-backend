package middleware

import (
	"net/http"
	"strings"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

const (
	// ContextUserClaimsKey is the key used to store token claims in the request context.
	ContextUserClaimsKey = "user_claims"
	// ContextCaregiverIDKey stores the caregiver identifier extracted from the token.
	ContextCaregiverIDKey = "caregiver_id"
)

// Authenticated ensures the incoming request presents a valid bearer token.
func Authenticated(authUC *usecase.AuthUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c)
			return
		}

		claims, err := authUC.ParseToken(parts[1])
		if err != nil {
			unauthorized(c)
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			unauthorized(c)
			return
		}

		c.Set(ContextUserClaimsKey, claims)
		c.Set(ContextCaregiverIDKey, sub)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":   domain.ErrUnauthorized.Error(),
		"message": "missing or invalid bearer token",
	})
}
