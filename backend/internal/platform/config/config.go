package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	HTTP        HTTP
	Database    Database
	Auth        AuthConfig
}

type HTTP struct {
	Address           string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	AllowedOrigins    []string
}

type Database struct {
	URL             string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}
type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment: env("APP_ENV", "development"),
		HTTP: HTTP{
			Address:        env("HTTP_ADDRESS", ":8080"),
			AllowedOrigins: splitCSV(env("HTTP_ALLOWED_ORIGINS", "http://localhost:3000")),
		},
		Database: Database{
			URL: env("DATABASE_URL", ""),
		},
		Auth: AuthConfig{
			JWTSecret: env("JWT_SECRET", ""),
		},
	}

	if strings.TrimSpace(cfg.Auth.JWTSecret) == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}

	var err error
	if cfg.HTTP.ReadTimeout, err = duration("HTTP_READ_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.ReadHeaderTimeout, err = duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.WriteTimeout, err = duration("HTTP_WRITE_TIMEOUT", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.IdleTimeout, err = duration("HTTP_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.ShutdownTimeout, err = duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.Database.ConnectTimeout, err = duration("DB_CONNECT_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.Database.MaxConnLifetime, err = duration("DB_MAX_CONN_LIFETIME", time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.Database.MaxConnIdleTime, err = duration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.Database.MaxConnections, err = int32Value("DB_MAX_CONNECTIONS", 20); err != nil {
		return Config{}, err
	}
	if cfg.Database.MinConnections, err = int32Value("DB_MIN_CONNECTIONS", 2); err != nil {
		return Config{}, err
	}
	if cfg.Database.MinConnections > cfg.Database.MaxConnections {
		return Config{}, fmt.Errorf("DB_MIN_CONNECTIONS must not exceed DB_MAX_CONNECTIONS")
	}
	if cfg.Auth.AccessTokenTTL, err = duration("ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.Auth.RefreshTokenTTL, err = duration("REFRESH_TOKEN_TTL", 720*time.Hour); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	value := env(key, fallback.String())
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return parsed, nil
}

func int32Value(key string, fallback int32) (int32, error) {
	value := env(key, strconv.FormatInt(int64(fallback), 10))
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%s must not be negative", key)
	}
	return int32(parsed), nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}
