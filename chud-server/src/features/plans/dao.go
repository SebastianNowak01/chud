package plans

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
)

type IPlanDAO interface {
	GetPlansByActivity(ctx context.Context, activityID string) ([]Plan, error)
	GetPlanByID(ctx context.Context, id string) (*Plan, error)
	InsertPlan(ctx context.Context, plan *Plan) (*Plan, error)
	UpdatePlan(ctx context.Context, plan *Plan) (*Plan, error)
	DeletePlan(ctx context.Context, id string) error
	GetLastScheduledFor(ctx context.Context, planID string) (*clock.Date, error)
}

type PlanDAO struct {
	pool db.Querier
}

func NewPlanDAO(pool db.Querier) IPlanDAO {
	return &PlanDAO{
		pool: pool,
	}
}

func (r *PlanDAO) GetPlansByActivity(ctx context.Context, activityID string) ([]Plan, error) {
	return db.Wrap(ctx, "GetPlansByActivity", func() ([]Plan, error) {
		plans := []Plan{}
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&plans,
			`SELECT * FROM plans WHERE activity_id = $1 ORDER BY created_at`,
			activityID,
		)
		return plans, err
	})
}

func (r *PlanDAO) GetPlanByID(ctx context.Context, id string) (*Plan, error) {
	return db.Wrap(ctx, "GetPlanByID", func() (*Plan, error) {
		return db.GetOne[Plan](ctx, r.pool, `SELECT * FROM plans WHERE id = $1`, id)
	})
}

func (r *PlanDAO) InsertPlan(ctx context.Context, plan *Plan) (*Plan, error) {
	return db.Wrap(ctx, "InsertPlan", func() (*Plan, error) {
		return db.GetOne[Plan](
			ctx,
			r.pool,
			`INSERT INTO plans (
				id, activity_id, user_id, title,
				monday, tuesday, wednesday, thursday, friday, saturday, sunday,
				starts_on, ends_on
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING *`,
			plan.ID, plan.ActivityID, plan.UserID, plan.Title,
			plan.Monday, plan.Tuesday, plan.Wednesday, plan.Thursday, plan.Friday, plan.Saturday, plan.Sunday,
			plan.StartsOn, plan.EndsOn,
		)
	})
}

func (r *PlanDAO) UpdatePlan(ctx context.Context, plan *Plan) (*Plan, error) {
	return db.Wrap(ctx, "UpdatePlan", func() (*Plan, error) {
		return db.GetOne[Plan](
			ctx,
			r.pool,
			`UPDATE plans SET title = $2, ends_on = $3 WHERE id = $1 RETURNING *`,
			plan.ID, plan.Title, plan.EndsOn,
		)
	})
}

func (r *PlanDAO) DeletePlan(ctx context.Context, id string) error {
	return db.WrapExec(ctx, "DeletePlan", r.pool, `DELETE FROM plans WHERE id = $1`, id)
}

func (r *PlanDAO) GetLastScheduledFor(ctx context.Context, planID string) (*clock.Date, error) {
	return db.Wrap(ctx, "GetLastScheduledFor", func() (*clock.Date, error) {
		var last *clock.Date
		err := sqlx.GetContext(ctx, r.pool, &last, `SELECT max(scheduled_for) FROM entries WHERE plan_id = $1`, planID)
		return last, err
	})
}
