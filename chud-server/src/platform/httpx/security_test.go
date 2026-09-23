package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Contains(t, rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
}

func TestPathUUIDOrRespond(t *testing.T) {
	mux := http.NewServeMux()
	var got string
	mux.HandleFunc("GET /things/{id}", func(w http.ResponseWriter, r *http.Request) {
		if id, ok := PathUUIDOrRespond(w, r, "id"); ok {
			got = id
		}
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/things/not-a-uuid", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)

	const id = "0b7c2a52-5f0e-4a8b-9d7e-3f1c2b4a6d8e"
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/things/"+id, nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, id, got)
}
