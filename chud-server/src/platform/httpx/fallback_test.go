package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/things", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := JSONFallback(mux)

	serve := func(method, path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		return rec
	}
	message := func(rec *httptest.ResponseRecorder) string {
		var body struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		return body.Error.Message
	}

	t.Run("matched route passes through", func(t *testing.T) {
		assert.Equal(t, http.StatusTeapot, serve(http.MethodGet, "/api/v1/things").Code)
	})

	t.Run("unknown path is a JSON 404", func(t *testing.T) {
		rec := serve(http.MethodGet, "/api/v1/nope")
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "nie znaleziono", message(rec))
	})

	t.Run("wrong method is a JSON 405", func(t *testing.T) {
		rec := serve(http.MethodDelete, "/api/v1/things")
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Contains(t, rec.Header().Get("Allow"), http.MethodGet)
		assert.Equal(t, "niedozwolona metoda", message(rec))
	})
}
