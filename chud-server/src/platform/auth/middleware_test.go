package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sebnow/chud/platform/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenFromRequest(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   string
	}{
		{http.MethodGet, "/api/v1/media/abc", "cookie-token"},
		{http.MethodGet, "/api/v1/me", ""},
		{http.MethodPost, "/api/v1/activities", ""},
		{http.MethodDelete, "/api/v1/media/abc", ""},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.AddCookie(&http.Cookie{Name: LoginTokenCookie, Value: "cookie-token"})
		assert.Equal(t, tc.want, tokenFromRequest(req), "%s %s", tc.method, tc.path)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/activities", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	assert.Equal(t, "header-token", tokenFromRequest(req))
}

func TestJwtAuthChecksSession(t *testing.T) {
	t.Setenv(config.JwtSecret, "test-secret")
	t.Setenv(config.JwtExpiryHours, "1")

	token, err := NewJwt("id-1", "alice", true, 1)
	require.NoError(t, err)

	cases := []struct {
		name    string
		session *Session
		status  int
		admin   bool
	}{
		{"current version", &Session{Version: 1, IsAdmin: true}, http.StatusOK, true},
		{"admin taken away", &Session{Version: 1}, http.StatusOK, false},
		{"password changed", &Session{Version: 2, IsAdmin: true}, http.StatusUnauthorized, false},
		{"user deleted", nil, http.StatusUnauthorized, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lookup := func(context.Context, string) (*Session, error) { return tc.session, nil }
			var admin bool
			handler := JwtAuth(lookup, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				claims, _ := GetUserClaimsFromContext(r.Context())
				admin = claims.IsAdmin
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.admin, admin)
		})
	}
}
