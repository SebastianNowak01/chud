package httpx

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebnow/chud/platform/log"
)

func Logger(ctx context.Context, next http.Handler) http.Handler {
	baseLogger := log.FromContext(ctx)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		logger := baseLogger.With().Str("request_id", reqID).Logger()
		r = r.WithContext(log.WithContext(r.Context(), &logger))

		shouldLog := strings.HasPrefix(r.URL.Path, "/api/")

		if shouldLog {
			logger.Trace().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("user_agent", r.UserAgent()).
				Str("remote_addr", r.RemoteAddr).
				Msg("Received request")
		}

		start := time.Now()

		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		if shouldLog {
			duration := time.Since(start).Truncate(1 * time.Microsecond)
			logger.Trace().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status_code", rw.statusCode).
				Dur("duration_ms", duration).
				Msg("Processed request")
		}
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				ctx := r.Context()
				log.FromContext(ctx).Error().Any("panic", p).Msg("Panic recovered")
				RespondError(ctx, w, http.StatusInternalServerError, errors.New("panic"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
