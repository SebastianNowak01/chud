package entries

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/platform/db"
)

const latestEntriesLimit = 50

type IEntryDAO interface {
	GetEntriesByActivity(ctx context.Context, activityID string) ([]Entry, error)
	GetEntryByID(ctx context.Context, id string) (*Entry, error)
	GetEntriesInRange(ctx context.Context, from, to time.Time) ([]Entry, error)
	GetUserEntriesInRange(ctx context.Context, userID string, from, to time.Time) ([]Entry, error)
	InsertEntry(ctx context.Context, entry *Entry) (*Entry, error)
	InsertMedia(ctx context.Context, entryID string, media []Media) error
	GetMediaByEntry(ctx context.Context, entryID string) ([]Media, error)
	GetMediaByID(ctx context.Context, id string) (*Media, error)
}

type EntryDAO struct {
	pool db.Querier
}

func NewEntryDAO(pool db.Querier) IEntryDAO {
	return &EntryDAO{
		pool: pool,
	}
}

func (r *EntryDAO) GetEntriesByActivity(ctx context.Context, activityID string) ([]Entry, error) {
	return db.Wrap(ctx, "GetEntriesByActivity", func() ([]Entry, error) {
		entries := []Entry{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&entries,
			`SELECT * FROM entries WHERE activity_id = $1 ORDER BY occurred_at DESC LIMIT $2`,
			activityID,
			latestEntriesLimit,
		)
		return entries, err
	})
}

func (r *EntryDAO) GetEntryByID(ctx context.Context, id string) (*Entry, error) {
	return db.Wrap(ctx, "GetEntryByID", func() (*Entry, error) {
		return db.GetOne[Entry](ctx, r.pool, `SELECT * FROM entries WHERE id = $1`, id)
	})
}

func (r *EntryDAO) GetEntriesInRange(ctx context.Context, from, to time.Time) ([]Entry, error) {
	return db.Wrap(ctx, "GetEntriesInRange", func() ([]Entry, error) {
		entries := []Entry{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&entries,
			`SELECT * FROM entries WHERE occurred_at >= $1 AND occurred_at < $2 ORDER BY occurred_at`,
			from,
			to,
		)
		return entries, err
	})
}

func (r *EntryDAO) GetUserEntriesInRange(ctx context.Context, userID string, from, to time.Time) ([]Entry, error) {
	return db.Wrap(ctx, "GetUserEntriesInRange", func() ([]Entry, error) {
		entries := []Entry{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&entries,
			`SELECT * FROM entries
			WHERE occurred_at >= $1 AND occurred_at < $2 AND user_id = $3
			ORDER BY occurred_at`,
			from,
			to,
			userID,
		)
		return entries, err
	})
}

func (r *EntryDAO) InsertEntry(ctx context.Context, entry *Entry) (*Entry, error) {
	return db.Wrap(ctx, "InsertEntry", func() (*Entry, error) {
		return db.GetOne[Entry](
			ctx,
			r.pool,
			`INSERT INTO entries (
				id, activity_id, user_id, plan_id, scheduled_for, excused, description, occurred_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING *`,
			entry.ID,
			entry.ActivityID,
			entry.UserID,
			entry.PlanID,
			entry.ScheduledFor,
			entry.Excused,
			entry.Description,
			entry.OccurredAt,
		)
	})
}

func (r *EntryDAO) InsertMedia(ctx context.Context, entryID string, media []Media) error {
	_, err := db.Wrap(ctx, "InsertMedia", func() (struct{}, error) {
		for _, m := range media {
			_, err := r.pool.ExecContext(
				ctx,
				`INSERT INTO media (id, entry_id, content_type, data) VALUES ($1, $2, $3, $4)`,
				m.ID,
				entryID,
				m.ContentType,
				m.Data,
			)
			if err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, nil
	})
	return err
}

func (r *EntryDAO) GetMediaByEntry(ctx context.Context, entryID string) ([]Media, error) {
	return db.Wrap(ctx, "GetMediaByEntry", func() ([]Media, error) {
		media := []Media{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&media,
			`SELECT id, entry_id, content_type FROM media WHERE entry_id = $1`,
			entryID,
		)
		return media, err
	})
}

func (r *EntryDAO) GetMediaByID(ctx context.Context, id string) (*Media, error) {
	return db.Wrap(ctx, "GetMediaByID", func() (*Media, error) {
		return db.GetOne[Media](ctx, r.pool, `SELECT * FROM media WHERE id = $1`, id)
	})
}
