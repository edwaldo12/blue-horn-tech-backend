package middleware

import (
	"strings"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/config"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/domain"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds a structured zap.Logger based on the logging configuration.
func NewLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Sampling:         nil,
		Encoding:         "json",
		EncoderConfig:    encoderCfg,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapCfg.Build()
}

// RequestLogger returns a Gin middleware that logs request metadata.
func RequestLogger(logger *zap.Logger, repo repository.RequestLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Set("request_start", start)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		logger.Info("request completed",
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Duration("latency", latency),
			zap.String("ip", clientIP),
			zap.String("userAgent", userAgent),
		)

		if repo != nil {
			logEntry := domain.RequestLog{
				Method:    method,
				Path:      path,
				Query:     raw,
				Status:    status,
				Latency:   latency,
				IP:        clientIP,
				UserAgent: userAgent,
			}
			if err := repo.Insert(c.Request.Context(), logEntry); err != nil {
				logger.Warn("failed to persist request log", zap.Error(err))
			}
		}
	}
}
