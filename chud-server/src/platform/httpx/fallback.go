package httpx

import (
	"errors"
	"net/http"
)

type statusRecorder struct {
	header http.Header
	status int
}

func (r *statusRecorder) Header() http.Header         { return r.header }
func (r *statusRecorder) Write(b []byte) (int, error) { return len(b), nil }
func (r *statusRecorder) WriteHeader(status int)      { r.status = status }

func JSONFallback(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}

		rec := &statusRecorder{header: http.Header{}, status: http.StatusNotFound}
		handler.ServeHTTP(rec, r)

		if rec.status == http.StatusMethodNotAllowed {
			w.Header().Set("Allow", rec.header.Get("Allow"))
			RespondError(r.Context(), w, http.StatusMethodNotAllowed, errors.New("niedozwolona metoda"))
			return
		}
		RespondError(r.Context(), w, http.StatusNotFound, errors.New("nie znaleziono"))
	})
}
