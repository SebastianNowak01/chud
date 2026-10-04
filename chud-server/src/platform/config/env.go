package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

const (
	APIPort        = "API_PORT"
	LogLevel       = "LOG_LEVEL"
	LogFormat      = "LOG_FORMAT"
	JwtSecret      = "JWT_SECRET"
	JwtExpiryHours = "JWT_EXPIRY_HOURS"
	AdminUser      = "ADMIN_USER"
	AdminPassword  = "ADMIN_PASSWORD"
	DatabaseURL    = "DATABASE_URL"
	AppTimezone    = "APP_TIMEZONE"
)

const minJwtSecretLength = 32

func Validate() error {
	log.Info().Msg("Validating environment variables...")
	var missingVars []string

	if os.Getenv(APIPort) == "" {
		os.Setenv(APIPort, "2137")
	}

	if os.Getenv(AppTimezone) == "" {
		os.Setenv(AppTimezone, "Europe/Warsaw")
	}

	if os.Getenv(JwtExpiryHours) == "" {
		os.Setenv(JwtExpiryHours, "24")
	}

	if os.Getenv(JwtSecret) == "" {
		missingVars = append(missingVars, JwtSecret)
	}

	if secret := os.Getenv(JwtSecret); secret != "" && len(secret) < minJwtSecretLength {
		return fmt.Errorf("%s must be at least %d bytes long", JwtSecret, minJwtSecretLength)
	}

	if os.Getenv(DatabaseURL) == "" {
		missingVars = append(missingVars, DatabaseURL)
	}

	if os.Getenv(AdminUser) == "" {
		missingVars = append(missingVars, AdminUser)
	}

	if os.Getenv(AdminPassword) == "" {
		missingVars = append(missingVars, AdminPassword)
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing environment variables: %v", missingVars)
	}

	return nil
}
