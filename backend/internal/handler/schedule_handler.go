package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// ScheduleHandler manages schedule endpoints.
type ScheduleHandler struct {
	scheduleUC *usecase.ScheduleUsecase
}

// NewScheduleHandler constructs the handler.
func NewScheduleHandler(scheduleUC *usecase.ScheduleUsecase) *ScheduleHandler {
	return &ScheduleHandler{scheduleUC: scheduleUC}
}

// ListSchedules returns schedules for the authenticated caregiver.
func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	filter := repository.ScheduleFilter{}
	if statusParam := c.Query("status"); statusParam != "" {
		statuses := strings.Split(statusParam, ",")
		for _, s := range statuses {
			filter.Status = append(filter.Status, domain.ScheduleStatus(strings.ToLower(strings.TrimSpace(s))))
		}
	}
	if dateStr := c.Query("date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			filter.Date = &parsed
		}
	}
	
	// Parse pagination parameters
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
	
	// Set default limit if not provided
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	summaries, err := h.scheduleUC.ListSchedules(c, caregiverID, filter)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	data := make([]gin.H, 0, len(summaries))
	for _, s := range summaries {
		data = append(data, gin.H{
			"id":            s.ID,
			"client_name":   s.ClientName,
			"service_name":  s.ServiceName,
			"start_time":    s.StartTime,
			"end_time":      s.EndTime,
			"status":        s.Status,
			"location_name": s.LocationName,
		})
	}

	// Include pagination info in response
	response := gin.H{
		"data": data,
		"pagination": gin.H{
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"hasMore": len(data) == filter.Limit, // If we got exactly limit items, there might be more
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetSchedule returns a single schedule detail.
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}
	scheduleID := c.Param("scheduleID")
	if scheduleID == "" {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, "missing schedule id")
		return
	}

	schedule, err := h.scheduleUC.GetSchedule(c, scheduleID, caregiverID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": scheduleToResponse(schedule),
	})
}

type clockEventRequest struct {
	Latitude  *float64 `json:"latitude" binding:"required"`
	Longitude *float64 `json:"longitude" binding:"required"`
	Notes     *string  `json:"notes"`
}

// StartSchedule logs clock in event.
func (h *ScheduleHandler) StartSchedule(c *gin.Context) {
	var req clockEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, err.Error())
		return
	}
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}
	scheduleID := c.Param("scheduleID")
	if scheduleID == "" {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, "missing schedule id")
		return
	}

	event := domain.VisitEvent{
		ScheduleID: scheduleID,
		Latitude:   *req.Latitude,
		Longitude:  *req.Longitude,
		Notes:      req.Notes,
	}

	schedule, err := h.scheduleUC.StartSchedule(c, scheduleID, caregiverID, event)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": scheduleToResponse(schedule),
	})
}

// EndSchedule logs clock out event.
func (h *ScheduleHandler) EndSchedule(c *gin.Context) {
	var req clockEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, err.Error())
		return
	}
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}
	scheduleID := c.Param("scheduleID")

	event := domain.VisitEvent{
		ScheduleID: scheduleID,
		Latitude:   *req.Latitude,
		Longitude:  *req.Longitude,
		Notes:      req.Notes,
	}
	schedule, err := h.scheduleUC.EndSchedule(c, scheduleID, caregiverID, event)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": scheduleToResponse(schedule),
	})
}

// TodaySchedules returns today's schedules plus metrics.
func (h *ScheduleHandler) TodaySchedules(c *gin.Context) {
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}
	today := time.Now().Format("2006-01-02")
	date, _ := time.Parse("2006-01-02", today)

	filter := repository.ScheduleFilter{Date: &date}
	summaries, err := h.scheduleUC.ListSchedules(c, caregiverID, filter)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	metrics, err := h.scheduleUC.GetMetrics(c, caregiverID, date)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	data := make([]gin.H, 0, len(summaries))
	for _, s := range summaries {
		data = append(data, gin.H{
			"id":            s.ID,
			"client_name":   s.ClientName,
			"service_name":  s.ServiceName,
			"start_time":    s.StartTime,
			"end_time":      s.EndTime,
			"status":        s.Status,
			"location_name": s.LocationName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"metrics": metrics,
	})
}

// Metrics returns aggregate schedule counts for a given date.
func (h *ScheduleHandler) Metrics(c *gin.Context) {
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	date := time.Now()
	if dateStr := c.Query("date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed
		}
	}

	metrics, err := h.scheduleUC.GetMetrics(c, caregiverID, date)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}

func scheduleToResponse(schedule domain.Schedule) gin.H {
	tasks := make([]gin.H, 0, len(schedule.Tasks))
	for _, t := range schedule.Tasks {
		tasks = append(tasks, gin.H{
			"id":                   t.ID,
			"title":                t.Title,
			"description":          t.Description,
			"status":               t.Status,
			"not_completed_reason": t.NotCompletedReason,
			"sort_order":           t.SortOrder,
			"updated_at":           t.UpdatedAt,
		})
	}

	return gin.H{
		"id":           schedule.ID,
		"caregiver_id": schedule.CaregiverID,
		"service_name": schedule.ServiceName,
		"client": gin.H{
			"id":        schedule.Client.ID,
			"full_name": schedule.Client.FullName,
			"email":     schedule.Client.Email,
			"phone":     schedule.Client.Phone,
			"address":   schedule.Client.Address,
			"city":      schedule.Client.City,
			"state":     schedule.Client.State,
			"postal":    schedule.Client.Postal,
			"latitude":  schedule.Client.Latitude,
			"longitude": schedule.Client.Longitude,
		},
		"start_time":     schedule.StartTime,
		"end_time":       schedule.EndTime,
		"status":         schedule.Status,
		"clock_in_at":    schedule.ClockInAt,
		"clock_in_lat":   schedule.ClockInLat,
		"clock_in_long":  schedule.ClockInLong,
		"clock_out_at":   schedule.ClockOutAt,
		"clock_out_lat":  schedule.ClockOutLat,
		"clock_out_long": schedule.ClockOutLong,
		"notes":          schedule.Notes,
		"tasks":          tasks,
		"location_label": schedule.LocationLabel,
		"duration_mins":  schedule.DurationMins,
	}
}
