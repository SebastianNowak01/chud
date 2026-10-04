package httpx

import (
	"net/http/httptest"
	"testing"

	"github.com/sebnow/chud/platform/config"
	"github.com/stretchr/testify/assert"
)

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 1.2.3.4")

	t.Setenv(config.RealIPHeader, "")
	assert.Equal(t, "10.0.0.1", ClientIP(req), "headers ignored unless configured")

	t.Setenv(config.RealIPHeader, "X-Forwarded-For")
	assert.Equal(t, "1.2.3.4", ClientIP(req), "last hop added by the proxy")

	t.Setenv(config.RealIPHeader, "X-Real-IP")
	assert.Equal(t, "10.0.0.1", ClientIP(req), "falls back when header is missing")
}
