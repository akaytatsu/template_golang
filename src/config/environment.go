package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var EnvironmentVariables EnvironmentVars

// ErrInvalidPortFormat is returned when a port environment variable cannot be parsed
var ErrInvalidPortFormat = errors.New("invalid port format: must be a number")

// ReadEnvironmentVars reads and validates all environment variables
// Returns an error if any required variable is missing or invalid
func ReadEnvironmentVars() error {
	EnvironmentVariables.ISRELEASE = os.Getenv("IS_RELEASE") == "true"

	// Logging configuration
	LogLevel := getEnvOrDefault("LOG_LEVEL", "INFO")
	GinMode := getEnvOrDefault("GIN_MODE", "release")
	GormLogLevel := getEnvOrDefault("GORM_LOG_LEVEL", "WARN")

	EnvironmentVariables.LogLevel = strings.ToUpper(LogLevel)
	EnvironmentVariables.GinMode = strings.ToLower(GinMode)
	EnvironmentVariables.GormLogLevel = strings.ToUpper(GormLogLevel)

	// Database configuration
	EnvironmentVariables.POSTGRES_DB = getEnvRequired("POSTGRES_DB")
	EnvironmentVariables.POSTGRES_USER = getEnvRequired("POSTGRES_USER")
	EnvironmentVariables.POSTGRES_PASSWORD = getEnvRequired("POSTGRES_PASSWORD")
	EnvironmentVariables.POSTGRES_HOST = getEnvRequired("POSTGRES_HOST")

	postgresPort, err := getEnvAsInt("POSTGRES_PORT", 5432)
	if err != nil {
		return fmt.Errorf("POSTGRES_PORT: %w", err)
	}
	EnvironmentVariables.POSTGRES_PORT = postgresPort

	// Kafka configuration
	EnvironmentVariables.KAFKA_BOOTSTRAP_SERVER = getEnvRequired("KAFKA_BOOTSTRAP_SERVER")
	EnvironmentVariables.KAFKA_CLIENT_ID = getEnvRequired("KAFKA_CLIENT_ID")
	EnvironmentVariables.KAFKA_GROUP_ID = getEnvRequired("KAFKA_GROUP_ID")

	// Email configuration
	EnvironmentVariables.EMAIL_HOST = getEnvOrDefault("EMAIL_HOST", "")
	EnvironmentVariables.EMAIL_HOST_USER = getEnvOrDefault("EMAIL_HOST_USER", "")
	EnvironmentVariables.EMAIL_HOST_PASSWORD = getEnvOrDefault("EMAIL_HOST_PASSWORD", "")

	emailPort, err := getEnvAsInt("EMAIL_PORT", 587)
	if err != nil {
		return fmt.Errorf("EMAIL_PORT: %w", err)
	}
	EnvironmentVariables.EMAIL_PORT = emailPort

	EnvironmentVariables.EMAIL_FROM = getEnvOrDefault("EMAIL_FROM", "")

	// Default admin configuration
	EnvironmentVariables.DEFAULT_ADMIN_MAIL = getEnvOrDefault("DEFAULT_ADMIN_MAIL", "")
	EnvironmentVariables.DEFAULT_ADMIN_PASSWORD = getEnvOrDefault("DEFAULT_ADMIN_PASSWORD", "")

	// JWT configuration - In production, JWT_SECRET_KEY must be set explicitly
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		if EnvironmentVariables.ISRELEASE {
			return errors.New("JWT_SECRET_KEY is required in production")
		}
		jwtSecret = "default-secret-key-change-in-production"
	}
	EnvironmentVariables.JWT_SECRET_KEY = jwtSecret

	return nil
}

// getEnvRequired gets an environment variable or panics
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

// getEnvOrDefault gets an environment variable or returns the default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer, with a default value
func getEnvAsInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidPortFormat, valueStr)
	}
	return value, nil
}
