package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

const (
	APIPort        = "API_PORT"
	LogLevel       = "LOG_LEVEL"  // TRACE, DEBUG, INFO, WARN, ERROR
	LogFormat      = "LOG_FORMAT" // JSON or CONSOLE
	JwtSecret      = "JWT_SECRET"
	JwtExpiryHours = "JWT_EXPIRY_HOURS"
	AdminUser      = "ADMIN_USER"
	AdminPassword  = "ADMIN_PASSWORD"
	DatabaseURL    = "DATABASE_URL"
)

func Validate() error {
	log.Info().Msg("Validating environment variables...")
	var missingVars []string

	if os.Getenv(APIPort) == "" {
		os.Setenv(APIPort, "2137")
	}

	if os.Getenv(JwtExpiryHours) == "" {
		os.Setenv(JwtExpiryHours, "24")
	}

	if os.Getenv(JwtSecret) == "" {
		missingVars = append(missingVars, JwtSecret)
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
