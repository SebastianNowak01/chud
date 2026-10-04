package summaries

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDown = errors.New("connection refused")

type fakeStats struct {
	mu          sync.Mutex
	down        bool
	occurrences []stats.Occurrence
}

func (f *fakeStats) GetOccurrences(_ context.Context, from, to clock.Date) ([]stats.Occurrence, *apperr.ServiceError) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.down {
		return nil, apperr.NewInternalError("%w", errDown)
	}
	var result []stats.Occurrence
	for _, o := range f.occurrences {
		if !o.Date.Before(from) && !o.Date.After(to) {
			result = append(result, o)
		}
	}
	return result, nil
}

func (f *fakeStats) GetLeaderboard(ctx context.Context, from, to clock.Date) ([]stats.UserStats, *apperr.ServiceError) {
	occurrences, svcErr := f.GetOccurrences(ctx, from, to)
	if svcErr != nil {
		return nil, svcErr
	}
	return stats.Leaderboard(testUsers, occurrences, nil), nil
}

type fakeUserDAO struct{ users.IUserDAO }

func (fakeUserDAO) GetAllUsers(context.Context) ([]users.User, error) { return testUsers, nil }

type fakeActivityDAO struct{ activities.IActivityDAO }

func (fakeActivityDAO) GetAllActivities(context.Context) ([]activities.Activity, error) {
	return testActivity, nil
}

type fakeSummaryDAO struct {
	mu     sync.Mutex
	down   bool
	stored map[clock.Date]Summary
}

func (f *fakeSummaryDAO) GetWeekEntries(context.Context, clock.Date, clock.Date) ([]entries.Entry, error) {
	return nil, nil
}

func (f *fakeSummaryDAO) GetSummary(_ context.Context, week clock.Date, _ int, now time.Time) (*Summary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.down {
		return nil, errDown
	}
	summary, ok := f.stored[week]
	if !ok || (summary.ExpiresAt != nil && !now.Before(*summary.ExpiresAt)) {
		return nil, db.ErrNotFound
	}
	return &summary, nil
}

func (f *fakeSummaryDAO) UpsertSummary(_ context.Context, summary *Summary) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.down {
		return errDown
	}
	f.stored[summary.Week] = *summary
	return nil
}

func (f *fakeSummaryDAO) DeleteExpiredSummaries(context.Context, time.Time) error { return nil }

type fakeLLM struct {
	mu       sync.Mutex
	disabled bool
	calls    int
	respond  func(call int) (string, error)
	gate     chan struct{}
}

func (f *fakeLLM) Enabled() bool { return !f.disabled }

func (f *fakeLLM) Complete(context.Context, string, string) (string, error) {
	f.mu.Lock()
	f.calls++
	call, respond, gate := f.calls, f.respond, f.gate
	f.mu.Unlock()
	if gate != nil {
		<-gate
	}
	if respond != nil {
		return respond(call)
	}
	return "Podsumowanie.", nil
}

func (f *fakeLLM) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type fixture struct {
	stats   *fakeStats
	dao     *fakeSummaryDAO
	llm     *fakeLLM
	service *SummaryService
}

var wednesdayNoon = d("2026-09-30").Start().Add(12 * time.Hour)

func newFixture() *fixture {
	f := &fixture{
		stats: &fakeStats{occurrences: []stats.Occurrence{
			occurrence("2026-09-21", "bob", "gym", stats.StatusDone),
			occurrence("2026-09-28", "alice", "gym", stats.StatusDone),
		}},
		dao: &fakeSummaryDAO{stored: map[clock.Date]Summary{}},
		llm: &fakeLLM{},
	}
	f.service = f.newService()
	f.at(wednesdayNoon)
	return f
}

func (f *fixture) newService() *SummaryService {
	return NewSummaryService(SummaryServiceDeps{
		SummaryDAO:   f.dao,
		StatsService: f.stats,
		UserDAO:      fakeUserDAO{},
		ActivityDAO:  fakeActivityDAO{},
		LLM:          f.llm,
		Clock:        clock.Fixed(wednesdayNoon),
	})
}

func (f *fixture) at(t time.Time) {
	c := clock.Fixed(t)
	f.service.clock.Store(&c)
}

func (f *fixture) get(t *testing.T, week clock.Date) *WeekSummary {
	t.Helper()
	view, svcErr := f.service.Get(context.Background(), week)
	require.Nil(t, svcErr)
	return view
}

func (f *fixture) generated(t *testing.T, week clock.Date) *WeekSummary {
	t.Helper()
	assert.Equal(t, StatusGenerating, f.get(t, week).Status)
	f.service.wg.Wait()
	return f.get(t, week)
}

func TestGetGeneratesOnceAndCaches(t *testing.T) {
	f := newFixture()

	view := f.generated(t, monday)
	assert.Equal(t, StatusReady, view.Status)
	assert.Equal(t, "Podsumowanie.", view.Text)
	assert.Equal(t, ptr(false), view.Outdated)
	assert.Equal(t, 1, f.llm.callCount())

	f.get(t, monday)
	assert.Equal(t, 1, f.llm.callCount(), "served from memory")

	stored := f.dao.stored[monday]
	assert.Equal(t, ptr(wednesdayNoon.Add(24*time.Hour)), stored.ExpiresAt)

	lastWeek := f.generated(t, d("2026-09-21"))
	assert.Equal(t, StatusReady, lastWeek.Status)
	assert.Nil(t, f.dao.stored[d("2026-09-21")].ExpiresAt, "past weeks are kept forever")
}

func TestConcurrentRequestsShareOneGeneration(t *testing.T) {
	f := newFixture()
	f.llm.gate = make(chan struct{})

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() { _, _ = f.service.Get(context.Background(), monday) })
	}
	wg.Wait()
	close(f.llm.gate)
	f.service.wg.Wait()

	assert.Equal(t, 1, f.llm.callCount())
	assert.Equal(t, StatusReady, f.get(t, monday).Status)
}

func TestStoredSummarySurvivesRestart(t *testing.T) {
	f := newFixture()
	f.generated(t, monday)

	f.service = f.newService()
	f.at(wednesdayNoon.Add(time.Hour))
	assert.Equal(t, StatusReady, f.get(t, monday).Status)
	assert.Equal(t, 1, f.llm.callCount())
}

func TestCurrentWeekExpiresAfterADay(t *testing.T) {
	f := newFixture()
	f.generated(t, monday)

	f.at(wednesdayNoon.Add(24 * time.Hour))
	assert.Equal(t, StatusGenerating, f.get(t, monday).Status)
	f.service.wg.Wait()
	assert.Equal(t, 2, f.llm.callCount())
}

func TestDatabaseDown(t *testing.T) {
	t.Run("memory still serves summaries", func(t *testing.T) {
		f := newFixture()
		f.generated(t, monday)

		f.stats.down, f.dao.down = true, true
		view := f.get(t, monday)
		assert.Equal(t, StatusReady, view.Status)
		assert.Nil(t, view.Outdated, "freshness unknown without data")
	})

	t.Run("nothing cached", func(t *testing.T) {
		f := newFixture()
		f.stats.down, f.dao.down = true, true
		assert.Equal(t, StatusUnavailable, f.get(t, monday).Status)
		assert.Equal(t, 0, f.llm.callCount())

		_, svcErr := f.service.Regenerate(context.Background(), monday)
		require.NotNil(t, svcErr)
		assert.Equal(t, http.StatusInternalServerError, svcErr.Code)
	})

	t.Run("store failure keeps summary in memory", func(t *testing.T) {
		f := newFixture()
		f.dao.down = true
		assert.Equal(t, StatusReady, f.generated(t, monday).Status)
		assert.Empty(t, f.dao.stored)
	})
}

func TestLLMFailure(t *testing.T) {
	f := newFixture()
	f.llm.respond = func(call int) (string, error) {
		if call == 1 {
			return "", errDown
		}
		return "Drugie podejście.", nil
	}

	assert.Equal(t, StatusUnavailable, f.generated(t, monday).Status)
	f.get(t, monday)
	assert.Equal(t, 1, f.llm.callCount(), "backs off after a failure")

	f.at(wednesdayNoon.Add(failureBackoff + time.Second))
	assert.Equal(t, "Drugie podejście.", f.generated(t, monday).Text)
}

func TestLLMPanicDoesNotCrash(t *testing.T) {
	f := newFixture()
	f.llm.respond = func(int) (string, error) { panic("boom") }
	assert.Equal(t, StatusUnavailable, f.generated(t, monday).Status)
}

func TestDisabledAndEmpty(t *testing.T) {
	f := newFixture()
	f.llm.disabled = true
	assert.Equal(t, StatusDisabled, f.get(t, monday).Status)
	_, svcErr := f.service.Regenerate(context.Background(), monday)
	assert.Equal(t, http.StatusConflict, svcErr.Code)

	f = newFixture()
	assert.Equal(t, StatusEmpty, f.get(t, d("2026-09-14")).Status)
	assert.Equal(t, 0, f.llm.callCount())
}

func TestOutdatedWhenDataChanges(t *testing.T) {
	f := newFixture()
	f.generated(t, monday)

	f.stats.mu.Lock()
	f.stats.occurrences = append(f.stats.occurrences, occurrence("2026-09-29", "bob", "gym", stats.StatusExcused))
	f.stats.mu.Unlock()

	view := f.get(t, monday)
	assert.Equal(t, StatusReady, view.Status)
	assert.Equal(t, ptr(true), view.Outdated)
	assert.Equal(t, 1, f.llm.callCount(), "outdated summaries are refreshed on demand only")
}

func TestRegenerateCooldown(t *testing.T) {
	f := newFixture()
	f.generated(t, monday)

	_, svcErr := f.service.Regenerate(context.Background(), monday)
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusTooManyRequests, svcErr.Code)

	f.at(wednesdayNoon.Add(regenerateWait))
	f.llm.respond = func(int) (string, error) { return "Nowe.", nil }
	view, svcErr := f.service.Regenerate(context.Background(), monday)
	require.Nil(t, svcErr)
	assert.Equal(t, StatusGenerating, view.Status)
	assert.Equal(t, "Podsumowanie.", view.Text, "old text stays visible while regenerating")
	f.service.wg.Wait()

	view = f.get(t, monday)
	assert.Equal(t, "Nowe.", view.Text)
	assert.Equal(t, ptr(wednesdayNoon.Add(2*regenerateWait)), view.RegenerateAvailableAt)
}

func TestValidateWeek(t *testing.T) {
	f := newFixture()
	for _, week := range []clock.Date{{}, d("2026-09-29"), d("2026-10-05")} {
		_, svcErr := f.service.Get(context.Background(), week)
		require.NotNil(t, svcErr, week.String())
		assert.Equal(t, http.StatusBadRequest, svcErr.Code)
	}
}
