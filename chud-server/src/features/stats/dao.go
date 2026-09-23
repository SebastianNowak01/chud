package stats

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
)

type IStatsDAO interface {
	GetPlansOverlapping(ctx context.Context, from, to clock.Date) ([]plans.Plan, error)
	GetPlannedEntries(ctx context.Context, from, to clock.Date) ([]PlannedEntry, error)
	CountUnplanned(ctx context.Context, from, to time.Time) ([]UnplannedCount, error)
}

type StatsDAO struct {
	pool db.Querier
}

func NewStatsDAO(pool db.Querier) IStatsDAO {
	return &StatsDAO{pool: pool}
}

func (r *StatsDAO) GetPlansOverlapping(ctx context.Context, from, to clock.Date) ([]plans.Plan, error) {
	return db.Wrap(ctx, "GetPlansOverlapping", func() ([]plans.Plan, error) {
		result := []plans.Plan{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&result,
			`SELECT * FROM plans WHERE starts_on <= $2 AND (ends_on IS NULL OR ends_on >= $1)`,
			from,
			to,
		)
		return result, err
	})
}

func (r *StatsDAO) GetPlannedEntries(ctx context.Context, from, to clock.Date) ([]PlannedEntry, error) {
	return db.Wrap(ctx, "GetPlannedEntries", func() ([]PlannedEntry, error) {
		result := []PlannedEntry{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&result,
			`SELECT id, plan_id, scheduled_for, excused FROM entries
			WHERE plan_id IS NOT NULL AND scheduled_for BETWEEN $1 AND $2`,
			from,
			to,
		)
		return result, err
	})
}

func (r *StatsDAO) CountUnplanned(ctx context.Context, from, to time.Time) ([]UnplannedCount, error) {
	return db.Wrap(ctx, "CountUnplanned", func() ([]UnplannedCount, error) {
		result := []UnplannedCount{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&result,
			`SELECT user_id, count(*) AS count FROM entries
			WHERE plan_id IS NULL AND occurred_at >= $1 AND occurred_at < $2
			GROUP BY user_id`,
			from,
			to,
		)
		return result, err
	})
}
