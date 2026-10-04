package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRequiresLongJwtSecret(t *testing.T) {
	t.Setenv(DatabaseURL, "postgres://x")
	t.Setenv(AdminUser, "admin")
	t.Setenv(AdminPassword, "admin-pass")

	t.Setenv(JwtSecret, "short")
	assert.Error(t, Validate())

	t.Setenv(JwtSecret, "dev-only-jwt-secret-change-me-in-prod")
	assert.NoError(t, Validate())
}
