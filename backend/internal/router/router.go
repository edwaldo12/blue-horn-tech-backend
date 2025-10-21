package router

import (
	"net/http"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/config"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/handler"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/middleware"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRouter wires handler endpoints and middleware.
func NewRouter(
	logger *zap.Logger,
	authUC *usecase.AuthUsecase,
	authHandler *handler.AuthHandler,
	scheduleHandler *handler.ScheduleHandler,
	taskHandler *handler.TaskHandler,
	attendanceHandler *handler.CaregiverAttendanceHandler,
	docsHandler *handler.DocsHandler,
	logRepo repository.RequestLogRepository,
	corsCfg config.CORSConfig,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger, logRepo))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsCfg.AllowOrigins,
		AllowMethods:     corsCfg.AllowMethods,
		AllowHeaders:     corsCfg.AllowHeaders,
		ExposeHeaders:    corsCfg.ExposeHeaders,
		AllowCredentials: corsCfg.AllowCredentials,
		MaxAge:           time.Duration(corsCfg.MaxAgeSeconds) * time.Second,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.GET("/docs", docsHandler.SwaggerUI)
	r.GET("/docs/openapi.yaml", docsHandler.OpenAPIYaml)

	api := r.Group("/api")
	{
		api.POST("/auth/token", authHandler.Token)
		protected := api.Group("/")
		protected.Use(middleware.Authenticated(authUC))

		protected.GET("/schedules", scheduleHandler.ListSchedules)
		protected.GET("/schedules/today", scheduleHandler.TodaySchedules)
		protected.GET("/schedules/metrics", scheduleHandler.Metrics)
		protected.GET("/schedules/:scheduleID", scheduleHandler.GetSchedule)
		protected.POST("/schedules/:scheduleID/start", scheduleHandler.StartSchedule)
		protected.POST("/schedules/:scheduleID/end", scheduleHandler.EndSchedule)
		protected.POST("/schedules/:scheduleID/tasks", taskHandler.CreateTask)

		protected.PATCH("/tasks/:taskID", taskHandler.UpdateTaskStatus)

		// Caregiver Attendance endpoints
		protected.GET("/attendance/today/status", attendanceHandler.GetTodayStatus)
		protected.POST("/attendance/clock-in", attendanceHandler.ClockIn)
		protected.POST("/attendance/clock-out", attendanceHandler.ClockOut)
		protected.GET("/attendance/history", attendanceHandler.GetAttendanceHistory)
	}

	return r
}
