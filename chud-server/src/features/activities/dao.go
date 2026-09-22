package activities

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/db"
)

type IActivityDAO interface {
	GetAllActivities(ctx context.Context) ([]Activity, error)
	GetActivityByID(ctx context.Context, id string) (*Activity, error)
	InsertActivity(ctx context.Context, activity *Activity) (*Activity, error)
	// GetMembers returns the users who posted at least one entry in the activity.
	GetMembers(ctx context.Context, activityID string) ([]users.User, error)
}

type ActivityDAO struct {
	pool db.Querier
}

func NewActivityDAO(pool db.Querier) IActivityDAO {
	return &ActivityDAO{
		pool: pool,
	}
}

func (r *ActivityDAO) GetAllActivities(ctx context.Context) ([]Activity, error) {
	return db.Wrap(ctx, "GetAllActivities", func() ([]Activity, error) {
		activities := []Activity{}
		err := sqlx.SelectContext(ctx, r.pool, &activities, `SELECT * FROM activities ORDER BY name`)
		return activities, err
	})
}

func (r *ActivityDAO) GetActivityByID(ctx context.Context, id string) (*Activity, error) {
	return db.Wrap(ctx, "GetActivityByID", func() (*Activity, error) {
		return db.GetOne[Activity](ctx, r.pool, `SELECT * FROM activities WHERE id = $1`, id)
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
