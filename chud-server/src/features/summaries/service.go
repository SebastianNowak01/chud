package summaries

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/cache"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/llm"
	"github.com/sebnow/chud/platform/log"
)

const (
	currentWeekTTL = 24 * time.Hour
	memoryTTL      = 7 * 24 * time.Hour
	regenerateWait = 5 * time.Minute
	failureBackoff = 5 * time.Minute
	storeTimeout   = 10 * time.Second
	maxCachedWeeks = 100
	maxQueued      = 4
	pruneThreshold = 256
)

type ISummaryService interface {
	Get(ctx context.Context, week clock.Date) (*WeekSummary, *apperr.ServiceError)
	Regenerate(ctx context.Context, week clock.Date) (*WeekSummary, *apperr.ServiceError)
}

type SummaryServiceDeps struct {
	Ctx          context.Context
	SummaryDAO   ISummaryDAO
	StatsService stats.IStatsService
	UserDAO      users.IUserDAO
	ActivityDAO  activities.IActivityDAO
	LLM          llm.Completer
	Clock        clock.Clock
}

type prepared struct {
	facts WeekFacts
	json  string
	hash  string
}

type SummaryService struct {
	ctx        context.Context
	dao        ISummaryDAO
	stats      stats.IStatsService
	userDAO    users.IUserDAO
	activities activities.IActivityDAO
	llm        llm.Completer
	clock      atomic.Pointer[clock.Clock]
	memory     *cache.LRU[clock.Date, Summary]
	slot       chan struct{}
	wg         sync.WaitGroup

	mu          sync.Mutex
	generating  map[clock.Date]bool
	startedAt   map[clock.Date]time.Time
	failedUntil map[clock.Date]time.Time
}

func NewSummaryService(deps SummaryServiceDeps) *SummaryService {
	ctx := deps.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	s := &SummaryService{
		ctx:         ctx,
		dao:         deps.SummaryDAO,
		stats:       deps.StatsService,
		userDAO:     deps.UserDAO,
		activities:  deps.ActivityDAO,
		llm:         deps.LLM,
		memory:      cache.New[clock.Date, Summary](maxCachedWeeks),
		slot:        make(chan struct{}, 1),
		generating:  map[clock.Date]bool{},
		startedAt:   map[clock.Date]time.Time{},
		failedUntil: map[clock.Date]time.Time{},
	}
	s.clock.Store(&deps.Clock)
	return s
}

func (s *SummaryService) now() time.Time {
	return s.clock.Load().Now()
}

func (s *SummaryService) Get(ctx context.Context, week clock.Date) (*WeekSummary, *apperr.ServiceError) {
	if svcErr := s.validateWeek(week); svcErr != nil {
		return nil, svcErr
	}
	if !s.llm.Enabled() {
		return &WeekSummary{Week: week, Status: StatusDisabled}, nil
	}

	input, factsErr := s.prepare(ctx, week)
	if factsErr != nil {
		log.FromContext(ctx).Warn().Err(factsErr).Str("week", week.String()).Msg("Week facts unavailable")
	}
	cached := s.cached(ctx, week)
	if cached == nil && input != nil && input.facts.Empty() {
		return &WeekSummary{Week: week, Status: StatusEmpty}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if cached == nil {
		if summary, ok := s.memory.Get(week); ok {
			cached = &summary
		}
	}
	if cached == nil && input != nil && !s.failedRecentlyLocked(week) {
		s.startLocked(week, input)
	}
	return s.viewLocked(week, cached, input), nil
}

func (s *SummaryService) Regenerate(ctx context.Context, week clock.Date) (*WeekSummary, *apperr.ServiceError) {
	if svcErr := s.validateWeek(week); svcErr != nil {
		return nil, svcErr
	}
	if !s.llm.Enabled() {
		return nil, apperr.NewConflictError("podsumowania są wyłączone")
	}
	input, err := s.prepare(ctx, week)
	if err != nil {
		return nil, apperr.NewInternalError("%w", err)
	}
	if input.facts.Empty() {
		return nil, apperr.NewConflictError("w tym tygodniu nie ma czego podsumować")
	}
	cached := s.cached(ctx, week)

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.generating[week] {
		if availableAt := s.regenerateAvailableAtLocked(week, cached); availableAt != nil && s.now().Before(*availableAt) {
			return nil, apperr.NewTooManyRequestsError("podsumowanie można odświeżyć raz na %d minut", int(regenerateWait.Minutes()))
		}
		delete(s.failedUntil, week)
		if !s.startLocked(week, input) {
			return nil, apperr.NewUnavailableError("model jest zajęty, spróbuj za chwilę")
		}
	}
	return s.viewLocked(week, cached, input), nil
}

func (s *SummaryService) validateWeek(week clock.Date) *apperr.ServiceError {
	if week.IsZero() || week.Weekday() != time.Monday {
		return apperr.NewBadRequestError("tydzień musi zaczynać się w poniedziałek")
	}
	if week.After(weekStart(clock.DateOf(s.now()))) {
		return apperr.NewBadRequestError("nie można podsumować przyszłego tygodnia")
	}
	return nil
}

func (s *SummaryService) prepare(ctx context.Context, week clock.Date) (*prepared, error) {
	end := week.AddDays(6)
	occurrences, svcErr := s.stats.GetOccurrences(ctx, week, end)
	if svcErr != nil {
		return nil, svcErr
	}
	board, svcErr := s.stats.GetLeaderboard(ctx, week, end)
	if svcErr != nil {
		return nil, svcErr
	}
	lastWeek, svcErr := s.stats.GetLeaderboard(ctx, week.AddDays(-7), week.AddDays(-1))
	if svcErr != nil {
		return nil, svcErr
	}
	userList, err := s.userDAO.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	activityList, err := s.activities.GetAllActivities(ctx)
	if err != nil {
		return nil, err
	}
	entryList, err := s.dao.GetWeekEntries(ctx, week, end)
	if err != nil {
		return nil, err
	}

	facts := BuildFacts(FactsInput{
		Week:        week,
		Today:       clock.DateOf(s.now()),
		Occurrences: occurrences,
		Leaderboard: board,
		LastWeek:    lastWeek,
		Users:       userList,
		Activities:  activityList,
		Entries:     entryList,
	})
	data, hash := facts.Encode()
	return &prepared{facts: facts, json: data, hash: hash}, nil
}

func (s *SummaryService) cached(ctx context.Context, week clock.Date) *Summary {
	now := s.now()
	if summary, ok := s.memory.Get(week); ok {
		if summary.ExpiresAt == nil || now.Before(*summary.ExpiresAt) {
			return &summary
		}
		s.memory.Delete(week)
	}
	summary, err := s.dao.GetSummary(ctx, week, PromptVersion, now)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			log.FromContext(ctx).Warn().Err(err).Str("week", week.String()).Msg("Stored summary unavailable")
		}
		return nil
	}
	s.remember(*summary, now)
	return summary
}

func (s *SummaryService) remember(summary Summary, now time.Time) {
	ttl := memoryTTL
	if summary.ExpiresAt != nil {
		ttl = summary.ExpiresAt.Sub(now)
	}
	if ttl > 0 {
		s.memory.Set(summary.Week, summary, ttl)
	}
}

func (s *SummaryService) failedRecentlyLocked(week clock.Date) bool {
	until, ok := s.failedUntil[week]
	return ok && s.now().Before(until)
}

func (s *SummaryService) regenerateAvailableAtLocked(week clock.Date, cached *Summary) *time.Time {
	last, ok := s.startedAt[week]
	if cached != nil && cached.GeneratedAt.After(last) {
		last, ok = cached.GeneratedAt, true
	}
	if !ok {
		return nil
	}
	availableAt := last.Add(regenerateWait)
	return &availableAt
}

func (s *SummaryService) startLocked(week clock.Date, input *prepared) bool {
	if s.generating[week] {
		return true
	}
	if len(s.generating) >= maxQueued {
		return false
	}
	now := s.now()
	s.pruneLocked(now)
	s.generating[week] = true
	s.startedAt[week] = now
	s.wg.Add(1)
	go s.generate(week, input)
	return true
}

func (s *SummaryService) pruneLocked(now time.Time) {
	if len(s.startedAt) < pruneThreshold && len(s.failedUntil) < pruneThreshold {
		return
	}
	for week, at := range s.startedAt {
		if now.Sub(at) > regenerateWait {
			delete(s.startedAt, week)
		}
	}
	for week, until := range s.failedUntil {
		if now.After(until) {
			delete(s.failedUntil, week)
		}
	}
}

func (s *SummaryService) generate(week clock.Date, input *prepared) {
	logger := log.FromContext(s.ctx)
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error().Interface("panic", r).Str("week", week.String()).Msg("Summary generation panicked")
			s.finish(week, nil)
		}
	}()

	select {
	case s.slot <- struct{}{}:
	case <-s.ctx.Done():
		s.finish(week, nil)
		return
	}
	defer func() { <-s.slot }()

	started := time.Now()
	text, err := s.llm.Complete(s.ctx, systemPrompt, userPrompt(input.json))
	if err == nil {
		text = cleanResponse(text)
		if text == "" {
			err = errors.New("llm returned only reasoning")
		}
	}
	if err != nil {
		logger.Warn().Err(err).Str("week", week.String()).Dur("duration_ms", time.Since(started)).Msg("Summary generation failed")
		s.finish(week, nil)
		return
	}
	logger.Info().Str("week", week.String()).Dur("duration_ms", time.Since(started)).Msg("Summary generated")

	now := s.now()
	summary := Summary{
		Week:          week,
		PromptVersion: PromptVersion,
		Text:          text,
		InputHash:     input.hash,
		GeneratedAt:   now,
		ExpiresAt:     weekExpiry(week, clock.DateOf(now), now),
	}
	s.remember(summary, now)
	s.store(summary, now)
	s.finish(week, &summary)
}

func (s *SummaryService) store(summary Summary, now time.Time) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), storeTimeout)
	defer cancel()
	logger := log.FromContext(s.ctx)
	if err := s.dao.UpsertSummary(ctx, &summary); err != nil {
		logger.Warn().Err(err).Str("week", summary.Week.String()).Msg("Summary kept only in memory")
		return
	}
	if err := s.dao.DeleteExpiredSummaries(ctx, now); err != nil {
		logger.Warn().Err(err).Msg("Expired summaries cleanup failed")
	}
}

func (s *SummaryService) finish(week clock.Date, summary *Summary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.generating, week)
	if summary == nil {
		s.failedUntil[week] = s.now().Add(failureBackoff)
	} else {
		delete(s.failedUntil, week)
	}
}

func (s *SummaryService) viewLocked(week clock.Date, cached *Summary, input *prepared) *WeekSummary {
	view := &WeekSummary{Week: week, Status: StatusUnavailable}
	if s.generating[week] {
		view.Status = StatusGenerating
	} else if cached != nil {
		view.Status = StatusReady
	}
	if cached != nil {
		generatedAt := cached.GeneratedAt
		view.Text = cached.Text
		view.GeneratedAt = &generatedAt
		if input != nil {
			outdated := cached.InputHash != input.hash
			view.Outdated = &outdated
		}
	}
	if !s.generating[week] && input != nil && !input.facts.Empty() {
		view.RegenerateAvailableAt = s.regenerateAvailableAtLocked(week, cached)
	}
	return view
}
