package handler

import (
	"net/http"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles task updates.
type TaskHandler struct {
	taskUC     *usecase.TaskUsecase
	scheduleUC *usecase.ScheduleUsecase
}

// NewTaskHandler constructs the handler.
func NewTaskHandler(taskUC *usecase.TaskUsecase, scheduleUC *usecase.ScheduleUsecase) *TaskHandler {
	return &TaskHandler{
		taskUC:     taskUC,
		scheduleUC: scheduleUC,
	}
}

type updateTaskRequest struct {
	ScheduleID string            `json:"schedule_id" binding:"required"`
	Status     domain.TaskStatus `json:"status" binding:"required"`
	Reason     *string           `json:"reason"`
}

// UpdateTaskStatus updates completion state for a task.
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	var req updateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, err.Error())
		return
	}
	caregiverID, ok := caregiverID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "")
		return
	}

	taskID := c.Param("taskID")
	if taskID == "" {
		respondError(c, http.StatusBadRequest, domain.ErrValidationFailure, "missing task id")
		return
	}

	if err := h.taskUC.UpdateTaskStatus(c, caregiverID, req.ScheduleID, taskID, req.Status, req.Reason); err != nil {
		handleDomainError(c, err)
		return
	}

	schedule, err := h.scheduleUC.GetSchedule(c, req.ScheduleID, caregiverID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scheduleToResponse(schedule)})
}

type createTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	SortOrder   int32  `json:"sort_order"`
}

// CreateTask appends a new task to the schedule.
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req createTaskRequest
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

	taskID, err := h.taskUC.AddTask(c, caregiverID, scheduleID, domain.Task{
		Title:       req.Title,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	schedule, err := h.scheduleUC.GetSchedule(c, scheduleID, caregiverID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    scheduleToResponse(schedule),
		"task_id": taskID,
	})
}
