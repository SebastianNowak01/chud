package summaries

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
	"github.com/stretchr/testify/assert"
)

type fakeService struct{ regenerateErr *apperr.ServiceError }

func (fakeService) Get(_ context.Context, week clock.Date) (*WeekSummary, *apperr.ServiceError) {
	return &WeekSummary{Week: week, Status: StatusReady}, nil
}

func (f fakeService) Regenerate(_ context.Context, week clock.Date) (*WeekSummary, *apperr.ServiceError) {
	if f.regenerateErr != nil {
		return nil, f.regenerateErr
	}
	return &WeekSummary{Week: week, Status: StatusGenerating}, nil
}

func TestSummaryRoutes(t *testing.T) {
	serve := func(service ISummaryService, method, target string) *httptest.ResponseRecorder {
		mux := http.NewServeMux()
		RegisterRoutes(mux, NewSummaryAPIController(service))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
		return rec
	}

	rec := serve(fakeService{}, http.MethodGet, "/api/v1/summaries/week?week=2026-09-28")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)

	assert.Equal(t, http.StatusBadRequest, serve(fakeService{}, http.MethodGet, "/api/v1/summaries/week?week=jutro").Code)
	assert.Equal(t, http.StatusAccepted, serve(fakeService{}, http.MethodPost, "/api/v1/summaries/week/regenerate?week=2026-09-28").Code)

	cooldown := fakeService{regenerateErr: apperr.NewTooManyRequestsError("za szybko")}
	assert.Equal(t, http.StatusTooManyRequests, serve(cooldown, http.MethodPost, "/api/v1/summaries/week/regenerate?week=2026-09-28").Code)
}
