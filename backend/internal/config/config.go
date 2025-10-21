package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Config aggregates environment backed configuration for the API.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	Logging   LoggingConfig
	CORS      CORSConfig
	Timezone  *time.Location
	StartTime time.Time
}

type AppConfig struct {
	Name string
	Env  string
	Host string
	Port int
}

type DatabaseConfig struct {
	Host         string
	Port         int
	Name         string
	User         string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

type AuthConfig struct {
	Issuer             string
	Audience           string
	ClientID           string
	ClientSecret       string
	AccessTokenSecret  string
	IDTokenSecret      string
	AccessTokenTTL     time.Duration
	IDTokenTTL         time.Duration
	DefaultCaregiverID string
}

type LoggingConfig struct {
	Level string
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

// Load constructs the Config struct by reading environment variables.
func Load() (Config, error) {
	appPort, err := getInt("APP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	dbPort, err := getInt("DB_PORT", 5432)
	if err != nil {
		return Config{}, err
	}

	dbMaxOpen, err := getInt("DB_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}

	dbMaxIdle, err := getInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}

	dbMaxIdleSeconds, err := getInt("DB_MAX_IDLE_TIME", 300)
	if err != nil {
		return Config{}, err
	}

	accessTTL, err := getDuration("AUTH_ACCESS_TOKEN_TTL", "15m")
	if err != nil {
		return Config{}, err
	}

	idTTL, err := getDuration("AUTH_ID_TOKEN_TTL", "60m")
	if err != nil {
		return Config{}, err
	}

	locationName := getString("APP_TIMEZONE", "UTC")
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return Config{}, fmt.Errorf("invalid APP_TIMEZONE %q: %w", locationName, err)
	}

	cfg := Config{
		App: AppConfig{
			Name: getString("APP_NAME", "care-shift-tracker"),
			Env:  getString("APP_ENV", "development"),
			Host: getString("APP_HOST", "0.0.0.0"),
			Port: appPort,
		},
		Database: DatabaseConfig{
			Host:         getString("DB_HOST", "localhost"),
			Port:         dbPort,
			Name:         getString("DB_NAME", "care_tracker"),
			User:         getString("DB_USER", "postgres"),
			Password:     getString("DB_PASSWORD", "postgres"),
			SSLMode:      getString("DB_SSL_MODE", "disable"),
			MaxOpenConns: dbMaxOpen,
			MaxIdleConns: dbMaxIdle,
			MaxIdleTime:  time.Duration(dbMaxIdleSeconds) * time.Second,
		},
		Auth: AuthConfig{
			Issuer:             getString("AUTH_ISSUER", "https://blue-horn-tech.local"),
			Audience:           getString("AUTH_AUDIENCE", "caregiver-app"),
			ClientID:           getString("AUTH_CLIENT_ID", "caregiver-app"),
			ClientSecret:       getString("AUTH_CLIENT_SECRET", "super-secret"),
			AccessTokenSecret:  getString("AUTH_ACCESS_TOKEN_SECRET", "access-secret-placeholder"),
			IDTokenSecret:      getString("AUTH_ID_TOKEN_SECRET", "id-secret-placeholder"),
			AccessTokenTTL:     accessTTL,
			IDTokenTTL:         idTTL,
			DefaultCaregiverID: getString("AUTH_DEFAULT_CAREGIVER_ID", ""),
		},
		Logging: LoggingConfig{
			Level: getString("LOG_LEVEL", "info"),
		},
		CORS: CORSConfig{
			AllowOrigins:     splitAndTrim(getString("CORS_ALLOW_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
			AllowMethods:     splitAndTrim(getString("CORS_ALLOW_METHODS", "GET,POST,PATCH,OPTIONS")),
			AllowHeaders:     splitAndTrim(getString("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization")),
			ExposeHeaders:    splitAndTrim(getString("CORS_EXPOSE_HEADERS", "Content-Length")),
			AllowCredentials: getBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAgeSeconds:    getIntOrFallback("CORS_MAX_AGE_SECONDS", 43200),
		},
		Timezone:  loc,
		StartTime: time.Now(),
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getInt(key string, fallback int) (int, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("invalid value for %s: %w", key, err)
		}
		return parsed, nil
	}
	return fallback, nil
}

func getIntOrFallback(key string, fallback int) int {
	val, err := getInt(key, fallback)
	if err != nil {
		return fallback
	}
	return val
}

func getDuration(key, fallback string) (time.Duration, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		d, err := time.ParseDuration(val)
		if err != nil {
			return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
		}
		return d, nil
	}
	return time.ParseDuration(fallback)
}

func getBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// NewPostgres connects to PostgreSQL using the provided configuration.
func NewPostgres(ctx context.Context, cfg DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.MaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
