package handler

import (
	"net/http"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// caregiverID extracts the caregiver id from the request context.
func caregiverID(c *gin.Context) (string, bool) {
	val, ok := c.Get(middleware.ContextCaregiverIDKey)
	if !ok {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// respondError standardises error payloads.
func respondError(c *gin.Context, status int, err error, message string) {
	if message == "" && err != nil {
		message = err.Error()
	}
	c.JSON(status, gin.H{
		"error":   err.Error(),
		"message": message,
	})
}

func handleDomainError(c *gin.Context, err error) {
	switch err {
	case nil:
		return
	case domain.ErrNotFound:
		respondError(c, http.StatusNotFound, err, "")
	case domain.ErrInvalidStatusTransition:
		respondError(c, http.StatusUnprocessableEntity, err, "")
	case domain.ErrValidationFailure:
		respondError(c, http.StatusBadRequest, err, "")
	case domain.ErrForbidden:
		respondError(c, http.StatusForbidden, err, "")
	default:
		respondError(c, http.StatusInternalServerError, err, "internal server error")
	}
}
