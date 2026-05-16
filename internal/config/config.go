package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	RedisAddr     string
	RedisPassword string
	RedisDB       string

	JwtSecret       string
	AccessTokenTTL  string
	RefreshTokenTTL string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "socialqueue"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "socialqueue"),
		PostgresDB:       getEnv("POSTGRES_DB", "socialqueue"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),

		JwtSecret:       getEnv("JWT_SECRET", "change-me-in-development"),
		AccessTokenTTL:  getEnv("ACCESS_TOKEN_TTL", "15m"),
		RefreshTokenTTL: getEnv("REFRESH_TOKEN_TTL", "720h"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
