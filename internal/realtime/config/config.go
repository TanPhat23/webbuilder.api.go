package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const defaultDevAllowedOrigins = "http://localhost:3000,https://webbuilderv2.vercel.app"

// Config holds all configuration for the application
type Config struct {
	ServerPort     string
	Database       DatabaseConfig
	Redis          RedisConfig
	DevMode        bool
	ClerkSecretKey string
	AllowedOrigins []string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string
	Host     string
	Port     string
	Password string
	DB       int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		ServerPort: envOrDefault("SERVER_PORT", "8080"),
		Database: DatabaseConfig{
			URL:      envOrDefault("DATABASE_URL", ""),
			Host:     envOrDefault("DB_HOST", "localhost"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "postgres"),
			Password: envOrDefault("DB_PASSWORD", "postgres"),
			Name:     envOrDefault("DB_NAME", "webbuilder"),
			SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			URL:      envOrDefault("REDIS_URL", ""),
			Host:     envOrDefault("REDIS_HOST", "localhost"),
			Port:     envOrDefault("REDIS_PORT", "6379"),
			Password: envOrDefault("REDIS_PASSWORD", ""),
			DB:       envIntOrDefault("REDIS_DB", 0),
		},
		DevMode:        envBoolOrDefault("DEV_MODE", false),
		ClerkSecretKey: envOrDefault("CLERK_SECRET_KEY", ""),
		AllowedOrigins: loadAllowedOrigins(),
	}, nil
}

func loadAllowedOrigins() []string {
	raw, ok := os.LookupEnv("ALLOWED_ORIGINS")
	if ok {
		return parseOrigins(raw)
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ENV")), "production") {
		return nil
	}
	return parseOrigins(defaultDevAllowedOrigins)
}

func parseOrigins(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return fallback
}
