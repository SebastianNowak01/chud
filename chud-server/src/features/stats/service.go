package stats

import (
	"context"

	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
)

type IStatsService interface {
	GetOccurrences(ctx context.Context, from, to clock.Date) ([]Occurrence, *apperr.ServiceError)
	GetLeaderboard(ctx context.Context, from, to clock.Date) ([]UserStats, *apperr.ServiceError)
}

type StatsServiceDeps struct {
	StatsDAO IStatsDAO
	UserDAO  users.IUserDAO
	Clock    clock.Clock
}

type StatsService struct {
	dao     IStatsDAO
	userDAO users.IUserDAO
	clock   clock.Clock
}

func NewStatsService(deps StatsServiceDeps) *StatsService {
	return &StatsService{dao: deps.StatsDAO, userDAO: deps.UserDAO, clock: deps.Clock}
}

func (s *StatsService) GetOccurrences(ctx context.Context, from, to clock.Date) ([]Occurrence, *apperr.ServiceError) {
	if err := clock.ValidateRange(from, to); err != nil {
		return nil, apperr.NewBadRequestError("%w", err)
	}
	planList, err := s.dao.GetPlansOverlapping(ctx, from, to)
	if err != nil {
		return nil, apperr.FromDAO(err, "plan")
	}
	entries, err := s.dao.GetPlannedEntries(ctx, from, to)
	if err != nil {
		return nil, apperr.FromDAO(err, "entry")
	}
	return Occurrences(planList, entries, from, to, s.clock.Today()), nil
}

func (s *StatsService) GetLeaderboard(ctx context.Context, from, to clock.Date) ([]UserStats, *apperr.ServiceError) {
	occurrences, svcErr := s.GetOccurrences(ctx, from, to)
	if svcErr != nil {
		return nil, svcErr
	}
	unplanned, err := s.dao.CountUnplanned(ctx, from.Start(), to.AddDays(1).Start())
	if err != nil {
		return nil, apperr.FromDAO(err, "entry")
	}
	userList, err := s.userDAO.GetAllUsers(ctx)
	if err != nil {
		return nil, apperr.FromDAO(err, "user")
	}
	return Leaderboard(userList, occurrences, unplanned), nil
}
