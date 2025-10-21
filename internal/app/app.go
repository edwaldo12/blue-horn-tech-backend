package app

import (
	"context"
	"fmt"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/config"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/handler"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/middleware"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository/postgres"
	routerpkg "github.com/edwaldo/test_blue_horn_tech/backend/internal/router"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Application wires core dependencies for the API server.
type Application struct {
	Config config.Config
	Logger *zap.Logger
	DB     *sqlx.DB
	Router *gin.Engine
}

// New builds the application container.
func New(ctx context.Context, cfg config.Config) (*Application, error) {
	log, err := middleware.NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	database, err := config.NewPostgres(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	schedRepo := postgres.NewScheduleRepository(database)
	taskRepo := postgres.NewTaskRepository(database)
	authRepo := postgres.NewAuthRepository(database)
	caregiverRepo := postgres.NewCaregiverRepository(database)
	caregiverLogRepo := postgres.NewCaregiverLogRepository(database)

	scheduleUC := usecase.NewScheduleUsecase(schedRepo, taskRepo)
	taskUC := usecase.NewTaskUsecase(taskRepo, schedRepo)
	authUC := usecase.NewAuthUsecase(cfg.Auth, authRepo, caregiverRepo)
	attendanceUC := usecase.NewCaregiverAttendanceUsecase(caregiverLogRepo)

	authHandler := handler.NewAuthHandler(authUC)
	scheduleHandler := handler.NewScheduleHandler(scheduleUC)
	taskHandler := handler.NewTaskHandler(taskUC, scheduleUC)
	attendanceHandler := handler.NewCaregiverAttendanceHandler(attendanceUC)
	docsHandler := handler.NewDocsHandler()

	router := routerpkg.NewRouter(log, authUC, authHandler, scheduleHandler, taskHandler, attendanceHandler, docsHandler, cfg.CORS)

	return &Application{
		Config: cfg,
		Logger: log,
		DB:     database,
		Router: router,
	}, nil
}

// Shutdown flushes resources cleanly.
func (a *Application) Shutdown(ctx context.Context) error {
	if err := a.DB.Close(); err != nil {
		return err
	}
	return a.Logger.Sync()
}
