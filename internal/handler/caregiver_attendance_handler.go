package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// CaregiverAttendanceHandler handles caregiver attendance endpoints.
type CaregiverAttendanceHandler struct {
	attendanceUC *usecase.CaregiverAttendanceUsecase
}

// NewCaregiverAttendanceHandler constructs the handler.
func NewCaregiverAttendanceHandler(attendanceUC *usecase.CaregiverAttendanceUsecase) *CaregiverAttendanceHandler {
	return &CaregiverAttendanceHandler{
		attendanceUC: attendanceUC,
	}
}

type attendanceClockRequest struct {
	Latitude  *float64 `json:"latitude" binding:"required"`
	Longitude *float64 `json:"longitude" binding:"required"`
	Notes     *string  `json:"notes"`
}

// ClockIn records a clock-in event for the authenticated caregiver.
func (h *CaregiverAttendanceHandler) ClockIn(c *gin.Context) {
	var req attendanceClockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, err.Error())
		return
	}

	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	log, err := h.attendanceUC.ClockIn(c, caregiverID, *req.Latitude, *req.Longitude, req.Notes)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": caregiverLogToResponse(log)})
}

// ClockOut records a clock-out event for the authenticated caregiver.
func (h *CaregiverAttendanceHandler) ClockOut(c *gin.Context) {
	var req attendanceClockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, err.Error())
		return
	}

	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	log, err := h.attendanceUC.ClockOut(c, caregiverID, *req.Latitude, *req.Longitude, req.Notes)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": caregiverLogToResponse(log)})
}

// GetTodayStatus returns today's attendance status for the authenticated caregiver.
func (h *CaregiverAttendanceHandler) GetTodayStatus(c *gin.Context) {
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	status, err := h.attendanceUC.GetTodayStatus(c, caregiverID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todayAttendanceStatusToResponse(status)})
}

// GetAttendanceHistory returns attendance history for the caregiver.
func (h *CaregiverAttendanceHandler) GetAttendanceHistory(c *gin.Context) {
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	filter := repository.CaregiverLogFilter{}
	
	// Parse log type
	if logTypeStr := c.Query("log_type"); logTypeStr != "" {
		logType := domain.LogType(logTypeStr)
		if logType == domain.LogTypeClockIn || logType == domain.LogTypeClockOut {
			filter.LogType = &logType
		}
	}
	
	// Parse date range
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	history, err := h.attendanceUC.GetAttendanceHistory(c, caregiverID, filter)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	result := make([]gin.H, len(history))
	for i, item := range history {
		result[i] = caregiverLogSummaryToResponse(item)
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func caregiverLogToResponse(log domain.CaregiverLog) gin.H {
	return gin.H{
		"id":           log.ID,
		"caregiver_id": log.CaregiverID,
		"log_type":     log.LogType,
		"latitude":     log.Latitude,
		"longitude":    log.Longitude,
		"timestamp":    log.Timestamp,
		"notes":        log.Notes,
		"created_at":   log.CreatedAt,
	}
}

func caregiverLogSummaryToResponse(log domain.CaregiverLogSummary) gin.H {
	return gin.H{
		"id":           log.ID,
		"caregiver_id": log.CaregiverID,
		"log_type":     log.LogType,
		"timestamp":    log.Timestamp,
	}
}

func todayAttendanceStatusToResponse(status usecase.TodayAttendanceStatus) gin.H {
	return gin.H{
		"caregiver_id":    status.CaregiverID,
		"date":            status.Date,
		"has_clocked_in":  status.HasClockedIn,
		"has_clocked_out": status.HasClockedOut,
		"clock_in_at":     status.ClockInAt,
		"clock_in_lat":    status.ClockInLat,
		"clock_in_long":   status.ClockInLong,
		"clock_out_at":    status.ClockOutAt,
		"clock_out_lat":   status.ClockOutLat,
		"clock_out_long":  status.ClockOutLong,
	}
}