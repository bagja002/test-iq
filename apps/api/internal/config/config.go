package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	Port               string
	FrontendURL        string
	MySQLDSN           string
	RedisURL           string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	CookieSecure       bool
	CookieDomain       string
	AutoRunMigrations  bool
	SeedDemoData       bool
	RateLimitMax       int
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBConnMaxLifetime  time.Duration
	SlowQueryThreshold time.Duration
}

func Load() *Config {
	appEnv := getString("APP_ENV", "development")
	seedDemoDefault := !strings.EqualFold(appEnv, "production")

	return &Config{
		AppEnv:             appEnv,
		Port:               getString("PORT", "8080"),
		FrontendURL:        strings.TrimRight(getString("FRONTEND_URL", "http://localhost:3000"), "/"),
		MySQLDSN:           getString("MYSQL_DSN", ""),
		RedisURL:           getString("REDIS_URL", ""),
		JWTSecret:          getString("JWT_SECRET", "change-this-secret"),
		AccessTokenTTL:     getDuration("ACCESS_TOKEN_TTL", 30*time.Minute),
		RefreshTokenTTL:    getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		CookieSecure:       getBool("COOKIE_SECURE", false),
		CookieDomain:       getString("COOKIE_DOMAIN", ""),
		AutoRunMigrations:  getBool("AUTO_RUN_MIGRATIONS", true),
		SeedDemoData:       getBool("SEED_DEMO_DATA", seedDemoDefault),
		RateLimitMax:       getInt("RATE_LIMIT_MAX", 15),
		DBMaxOpenConns:     getInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:     getInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime:  getDuration("DB_CONN_MAX_LIFETIME", 15*time.Minute),
		SlowQueryThreshold: getDuration("SLOW_QUERY_THRESHOLD", 250*time.Millisecond),
	}
}

func getString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
