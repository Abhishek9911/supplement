package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort      string
	DatabaseURL     string
	JWTSecret       string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration
	AllowedOrigins  []string
	AppEnv          string
	RateLimitWindow time.Duration
	RateLimitMax    int
}

func Load() Config {
	ttlAccess, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	ttlRefresh, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	window, _ := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	max, _ := strconv.Atoi(getEnv("RATE_LIMIT_MAX", "120"))
	allowedOrigins := parseList(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"))

	return Config{
		ServerPort:      getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/suppliment"),
		JWTSecret:       getEnv("JWT_SECRET", "change-this-super-secret-key"),
		JWTAccessTTL:    ttlAccess,
		JWTRefreshTTL:   ttlRefresh,
		AllowedOrigins:  allowedOrigins,
		AppEnv:          getEnv("APP_ENV", "development"),
		RateLimitWindow: window,
		RateLimitMax:    max,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func parseList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trim := strings.TrimSpace(part)
		if trim != "" {
			out = append(out, trim)
		}
	}
	return out
}
