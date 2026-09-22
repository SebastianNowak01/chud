package plans

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeActivityDAO struct{ activities.IActivityDAO }

func (fakeActivityDAO) GetActivityByID(_ context.Context, id string) (*activities.Activity, error) {
	if id != "gym" {
		return nil, db.ErrNotFound
	}
	return &activities.Activity{ID: id}, nil
}

func (fakeActivityDAO) GetMembers(context.Context, string) ([]users.User, error) { return nil, nil }

type fakePlanDAO struct{ plans map[string]Plan }

func (d *fakePlanDAO) GetPlansByActivity(context.Context, string) ([]Plan, error) { return nil, nil }

func (d *fakePlanDAO) GetPlanByID(_ context.Context, id string) (*Plan, error) {
	p, ok := d.plans[id]
	if !ok {
		return nil, db.ErrNotFound
	}
	return &p, nil
}

func (d *fakePlanDAO) InsertPlan(_ context.Context, p *Plan) (*Plan, error) {
	d.plans[p.ID] = *p
	return p, nil
}

func (d *fakePlanDAO) UpdatePlan(_ context.Context, p *Plan) (*Plan, error) {
	d.plans[p.ID] = *p
	return p, nil
}

func (d *fakePlanDAO) DeletePlan(_ context.Context, id string) error {
	delete(d.plans, id)
	return nil
}

func newTestService() *PlanService {
	return NewPlanService(PlanServiceDeps{
		PlanDAO:     &fakePlanDAO{plans: map[string]Plan{}},
		ActivityDAO: fakeActivityDAO{},
	})
}

func validPayload() PlanPayload {
	return PlanPayload{Title: "Gym 3x", Monday: true, Wednesday: true, Friday: true, StartsOn: "2026-09-21"}
}

func TestCreatePlan(t *testing.T) {
	svc := newTestService()

	plan, svcErr := svc.CreatePlan(context.Background(), "gym", "alice", validPayload())
	require.Nil(t, svcErr)
	assert.Equal(t, "Gym 3x", plan.Title)
	assert.True(t, plan.Monday)
	assert.False(t, plan.Tuesday)
	assert.Nil(t, plan.EndsOn)
}

func TestCreatePlanValidation(t *testing.T) {
	svc := newTestService()
	endsBeforeStart := "2026-09-01"

	tests := []struct {
		name     string
		activity string
		mutate   func(p *PlanPayload)
		wantCode int
	}{
		{"unknown activity", "nope", func(*PlanPayload) {}, http.StatusNotFound},
		{"empty title", "gym", func(p *PlanPayload) { p.Title = " " }, http.StatusBadRequest},
		{"no days", "gym", func(p *PlanPayload) { p.Monday, p.Wednesday, p.Friday = false, false, false }, http.StatusBadRequest},
		{"bad date", "gym", func(p *PlanPayload) { p.StartsOn = "21.09.2026" }, http.StatusBadRequest},
		{"ends before start", "gym", func(p *PlanPayload) { p.EndsOn = &endsBeforeStart }, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := validPayload()
			tt.mutate(&payload)
			_, svcErr := svc.CreatePlan(context.Background(), tt.activity, "alice", payload)
			require.NotNil(t, svcErr)
			assert.Equal(t, tt.wantCode, svcErr.Code)
		})
	}
}

func TestOnlyOwnerChangesPlan(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	plan, svcErr := svc.CreatePlan(ctx, "gym", "alice", validPayload())
	require.Nil(t, svcErr)

	_, svcErr = svc.UpdatePlan(ctx, plan.ID, "bob", validPayload())
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusForbidden, svcErr.Code)

	svcErr = svc.DeletePlan(ctx, plan.ID, "bob")
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusForbidden, svcErr.Code)

	require.Nil(t, svc.DeletePlan(ctx, plan.ID, "alice"))
}

func TestIsScheduledOn(t *testing.T) {
	day := func(s string) time.Time {
		d, err := time.Parse(time.DateOnly, s)
		require.NoError(t, err)
		return d
	}
	endsOn := day("2026-09-30")
	plan := Plan{Monday: true, Friday: true, StartsOn: day("2026-09-21"), EndsOn: &endsOn}

	assert.True(t, plan.IsScheduledOn(day("2026-09-21")), "first Monday")
	assert.True(t, plan.IsScheduledOn(day("2026-09-25")), "Friday")
	assert.False(t, plan.IsScheduledOn(day("2026-09-23")), "Wednesday is not planned")
	assert.False(t, plan.IsScheduledOn(day("2026-09-14")), "before start")
	assert.False(t, plan.IsScheduledOn(day("2026-10-02")), "after end")
}
