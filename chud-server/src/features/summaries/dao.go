package summaries

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
)

type ISummaryDAO interface {
	GetWeekEntries(ctx context.Context, from, to clock.Date) ([]entries.Entry, error)
	GetSummary(ctx context.Context, week clock.Date, promptVersion int, now time.Time) (*Summary, error)
	UpsertSummary(ctx context.Context, summary *Summary) error
	DeleteExpiredSummaries(ctx context.Context, now time.Time) error
}

type SummaryDAO struct {
	pool db.Querier
}

func NewSummaryDAO(pool db.Querier) ISummaryDAO {
	return &SummaryDAO{pool: pool}
}

func (r *SummaryDAO) GetWeekEntries(ctx context.Context, from, to clock.Date) ([]entries.Entry, error) {
	return db.Wrap(ctx, "GetWeekEntries", func() ([]entries.Entry, error) {
		result := []entries.Entry{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&result,
			`SELECT * FROM entries
			WHERE (plan_id IS NULL AND occurred_at >= $1 AND occurred_at < $2)
				OR scheduled_for BETWEEN $3 AND $4
			ORDER BY occurred_at, id`,
			from.Start(),
			to.AddDays(1).Start(),
			from,
			to,
		)
		return result, err
	})
}

func (r *SummaryDAO) GetSummary(ctx context.Context, week clock.Date, promptVersion int, now time.Time) (*Summary, error) {
	return db.Wrap(ctx, "GetSummary", func() (*Summary, error) {
		return db.GetOne[Summary](
			ctx,
			r.pool,
			`SELECT * FROM week_summaries
			WHERE week = $1 AND prompt_version = $2 AND (expires_at IS NULL OR expires_at > $3)`,
			week,
			promptVersion,
			now,
		)
	})
}

func (r *SummaryDAO) UpsertSummary(ctx context.Context, summary *Summary) error {
	return db.WrapExec(
		ctx,
		"UpsertSummary",
		r.pool,
		`INSERT INTO week_summaries (week, prompt_version, text, input_hash, generated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (week, prompt_version) DO UPDATE SET
			text = EXCLUDED.text,
			input_hash = EXCLUDED.input_hash,
			generated_at = EXCLUDED.generated_at,
			expires_at = EXCLUDED.expires_at`,
		summary.Week,
		summary.PromptVersion,
		summary.Text,
		summary.InputHash,
		summary.GeneratedAt,
		summary.ExpiresAt,
	)
}

func (r *SummaryDAO) DeleteExpiredSummaries(ctx context.Context, now time.Time) error {
	_, err := db.Wrap(ctx, "DeleteExpiredSummaries", func() (struct{}, error) {
		_, err := r.pool.ExecContext(ctx, `DELETE FROM week_summaries WHERE expires_at <= $1`, now)
		return struct{}{}, err
	})
	return err
}
