package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
