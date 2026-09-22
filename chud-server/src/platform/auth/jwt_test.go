package auth

import (
	"testing"

	"github.com/sebnow/chud/platform/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtRoundTrip(t *testing.T) {
	t.Setenv(config.JwtSecret, "test-secret")
	t.Setenv(config.JwtExpiryHours, "1")

	token, err := NewJwt("id-1", "admin", true)
	require.NoError(t, err)

	claims, err := ValidateJwt(token)
	require.NoError(t, err)
	assert.Equal(t, "id-1", claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.True(t, claims.IsAdmin)
}

func TestJwtRejectsWrongSecret(t *testing.T) {
	t.Setenv(config.JwtSecret, "test-secret")
	t.Setenv(config.JwtExpiryHours, "1")

	token, err := NewJwt("id-1", "admin", true)
	require.NoError(t, err)

	t.Setenv(config.JwtSecret, "other-secret")
	_, err = ValidateJwt(token)
	assert.Error(t, err)
}

func TestJwtRejectsExpired(t *testing.T) {
	t.Setenv(config.JwtSecret, "test-secret")
	t.Setenv(config.JwtExpiryHours, "-1")

	token, err := NewJwt("id-1", "admin", true)
	require.NoError(t, err)

	_, err = ValidateJwt(token)
	assert.Error(t, err)
}
