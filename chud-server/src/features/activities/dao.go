package activities

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
)

type IActivityDAO interface {
	GetAllActivities(ctx context.Context) ([]Activity, error)
	GetActivityByID(ctx context.Context, id string) (*Activity, error)
	GetActivityByName(ctx context.Context, name string) (*Activity, error)
	InsertActivity(ctx context.Context, activity *Activity) (*Activity, error)
	GetMembers(ctx context.Context, activityID string) ([]users.User, error)
	DeleteEmptyActivity(ctx context.Context, id string) error
	ArchiveActivity(ctx context.Context, id string, today clock.Date) (*Activity, error)
	RestoreActivity(ctx context.Context, id string) (*Activity, error)
}

type ActivityDAO struct {
	pool db.Querier
}

func NewActivityDAO(pool db.Querier) IActivityDAO {
	return &ActivityDAO{
		pool: pool,
	}
}

const selectActivity = `SELECT a.*,
	EXISTS (SELECT 1 FROM plans p WHERE p.activity_id = a.id)
		OR EXISTS (SELECT 1 FROM entries e WHERE e.activity_id = a.id) AS has_history
	FROM activities a`

func (r *ActivityDAO) GetAllActivities(ctx context.Context) ([]Activity, error) {
	return db.Wrap(ctx, "GetAllActivities", func() ([]Activity, error) {
		activities := []Activity{}
		err := sqlx.SelectContext(ctx, r.pool, &activities, selectActivity+` ORDER BY a.name`)
		return activities, err
	})
}

func (r *ActivityDAO) GetActivityByID(ctx context.Context, id string) (*Activity, error) {
	return db.Wrap(ctx, "GetActivityByID", func() (*Activity, error) {
		return db.GetOne[Activity](ctx, r.pool, selectActivity+` WHERE a.id = $1`, id)
	})
}

func (r *ActivityDAO) GetActivityByName(ctx context.Context, name string) (*Activity, error) {
	return db.Wrap(ctx, "GetActivityByName", func() (*Activity, error) {
		return db.GetOne[Activity](ctx, r.pool, selectActivity+` WHERE a.name = $1`, name)
	})
}

func (r *ActivityDAO) InsertActivity(ctx context.Context, activity *Activity) (*Activity, error) {
	return db.Wrap(ctx, "InsertActivity", func() (*Activity, error) {
		return db.GetOne[Activity](
			ctx,
			r.pool,
			`INSERT INTO activities (id, name, description, created_by)
			VALUES ($1, $2, $3, $4)
			RETURNING *`,
			activity.ID,
			activity.Name,
			activity.Description,
			activity.CreatedBy,
		)
	})
}

func (r *ActivityDAO) GetMembers(ctx context.Context, activityID string) ([]users.User, error) {
	return db.Wrap(ctx, "GetMembers", func() ([]users.User, error) {
		members := []users.User{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&members,
			`SELECT * FROM users
			WHERE id IN (SELECT user_id FROM entries WHERE activity_id = $1)
			ORDER BY username`,
			activityID,
		)
		return members, err
	})
}

func (r *ActivityDAO) DeleteEmptyActivity(ctx context.Context, id string) error {
	return db.WrapExec(ctx, "DeleteEmptyActivity", r.pool,
		`DELETE FROM activities a
		WHERE a.id = $1
			AND NOT EXISTS (SELECT 1 FROM plans p WHERE p.activity_id = a.id)
			AND NOT EXISTS (SELECT 1 FROM entries e WHERE e.activity_id = a.id)`,
		id,
	)
}

func (r *ActivityDAO) ArchiveActivity(ctx context.Context, id string, today clock.Date) (*Activity, error) {
	return db.Wrap(ctx, "ArchiveActivity", func() (*Activity, error) {
		return db.GetOne[Activity](
			ctx,
			r.pool,
			`WITH target AS (
				SELECT id FROM activities WHERE id = $1 AND archived_at IS NULL FOR UPDATE
			),
			open_plans AS (
				SELECT p.id, p.starts_on,
					(SELECT max(e.scheduled_for) FROM entries e WHERE e.plan_id = p.id) AS last_entry
				FROM plans p
				WHERE p.activity_id IN (SELECT id FROM target)
					AND (p.ends_on IS NULL OR p.ends_on >= $2::date)
			),
			removed AS (
				DELETE FROM plans
				WHERE id IN (SELECT id FROM open_plans WHERE starts_on >= $2::date AND last_entry IS NULL)
			),
			ended AS (
				UPDATE plans p
				SET ends_on = GREATEST($2::date - 1, o.last_entry, o.starts_on)
				FROM open_plans o
				WHERE p.id = o.id AND NOT (o.starts_on >= $2::date AND o.last_entry IS NULL)
			)
			UPDATE activities SET archived_at = NOW()
			WHERE id IN (SELECT id FROM target)
			RETURNING *, TRUE AS has_history`,
			id,
			today,
		)
	})
}

func (r *ActivityDAO) RestoreActivity(ctx context.Context, id string) (*Activity, error) {
	return db.Wrap(ctx, "RestoreActivity", func() (*Activity, error) {
		return db.GetOne[Activity](
			ctx,
			r.pool,
			`UPDATE activities a SET archived_at = NULL
			WHERE a.id = $1
			RETURNING a.*,
				EXISTS (SELECT 1 FROM plans p WHERE p.activity_id = a.id)
					OR EXISTS (SELECT 1 FROM entries e WHERE e.activity_id = a.id) AS has_history`,
			id,
		)
	})
}
