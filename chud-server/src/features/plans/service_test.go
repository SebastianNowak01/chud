package plans

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
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

type fakePlanDAO struct {
	plans map[string]Plan
	last  map[string]clock.Date
}

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

func (d *fakePlanDAO) GetLastScheduledFor(_ context.Context, planID string) (*clock.Date, error) {
	last, ok := d.last[planID]
	if !ok {
		return nil, nil
	}
	return &last, nil
}

func day(s string) clock.Date {
	d, err := clock.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func at(s string) clock.Clock {
	return clock.Fixed(day(s).Start().Add(12 * time.Hour))
}

func newTestService() (*PlanService, *fakePlanDAO) {
	dao := &fakePlanDAO{plans: map[string]Plan{}, last: map[string]clock.Date{}}
	return NewPlanService(PlanServiceDeps{
		PlanDAO:     dao,
		ActivityDAO: fakeActivityDAO{},
		Clock:       at("2026-09-21"),
	}), dao
}

func validPayload() PlanPayload {
	return PlanPayload{Title: "Gym 3x", Monday: true, Wednesday: true, Friday: true, StartsOn: day("2026-09-21")}
}

func TestCreatePlan(t *testing.T) {
	svc, _ := newTestService()

	plan, svcErr := svc.CreatePlan(context.Background(), "gym", "alice", validPayload())
	require.Nil(t, svcErr)
	assert.Equal(t, "Gym 3x", plan.Title)
	assert.True(t, plan.Monday)
	assert.False(t, plan.Tuesday)
	assert.Nil(t, plan.EndsOn)
}

func TestCreatePlanValidation(t *testing.T) {
	svc, _ := newTestService()
	endsBeforeStart := day("2026-09-01")

	tests := []struct {
		name     string
		activity string
		mutate   func(p *PlanPayload)
		wantCode int
	}{
		{"unknown activity", "nope", func(*PlanPayload) {}, http.StatusNotFound},
		{"empty title", "gym", func(p *PlanPayload) { p.Title = " " }, http.StatusBadRequest},
		{"no days", "gym", func(p *PlanPayload) { p.Monday, p.Wednesday, p.Friday = false, false, false }, http.StatusBadRequest},
		{"missing start", "gym", func(p *PlanPayload) { p.StartsOn = clock.Date{} }, http.StatusBadRequest},
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

func TestCreatePlanCannotStartInThePast(t *testing.T) {
	svc, _ := newTestService()
	payload := validPayload()
	payload.StartsOn = day("2026-09-20")

	_, svcErr := svc.CreatePlan(context.Background(), "gym", "alice", payload)
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusBadRequest, svcErr.Code)
}

func TestOnlyOwnerChangesPlan(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestService()
	plan, svcErr := svc.CreatePlan(ctx, "gym", "alice", validPayload())
	require.Nil(t, svcErr)

	title := "Gym 4x"
	_, svcErr = svc.UpdatePlan(ctx, plan.ID, "bob", PlanUpdatePayload{Title: &title})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusForbidden, svcErr.Code)

	svcErr = svc.DeletePlan(ctx, plan.ID, "bob")
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusForbidden, svcErr.Code)

	require.Nil(t, svc.DeletePlan(ctx, plan.ID, "alice"))
}

func TestUpdatePlanKeepsSchedule(t *testing.T) {
	ctx := context.Background()
	svc, dao := newTestService()
	plan, svcErr := svc.CreatePlan(ctx, "gym", "alice", validPayload())
	require.Nil(t, svcErr)
	svc.clock = at("2026-10-10")
	dao.last[plan.ID] = day("2026-10-12")

	str := func(s string) *string { return &s }
	tests := []struct {
		name     string
		payload  PlanUpdatePayload
		wantCode int
	}{
		{"rename", PlanUpdatePayload{Title: str(" Gym 4x ")}, 0},
		{"empty title", PlanUpdatePayload{Title: str(" ")}, http.StatusBadRequest},
		{"end before yesterday", PlanUpdatePayload{EndsOn: str("2026-10-08")}, http.StatusBadRequest},
		{"end before last entry", PlanUpdatePayload{EndsOn: str("2026-10-11")}, http.StatusBadRequest},
		{"end before start", PlanUpdatePayload{EndsOn: str("2026-09-20")}, http.StatusBadRequest},
		{"end after last entry", PlanUpdatePayload{EndsOn: str("2026-10-12")}, 0},
		{"reopen", PlanUpdatePayload{EndsOn: str("")}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, svcErr := svc.UpdatePlan(ctx, plan.ID, "alice", tt.payload)
			if tt.wantCode != 0 {
				require.NotNil(t, svcErr)
				assert.Equal(t, tt.wantCode, svcErr.Code)
				return
			}
			require.Nil(t, svcErr)
			assert.True(t, updated.Monday && updated.Wednesday && updated.Friday)
			assert.Equal(t, day("2026-09-21"), updated.StartsOn)
		})
	}

	got, _ := dao.GetPlanByID(ctx, plan.ID)
	assert.Equal(t, "Gym 4x", got.Title)
	assert.Nil(t, got.EndsOn)
}

func TestDeletePlanOnlyWithoutHistory(t *testing.T) {
	ctx := context.Background()
	svc, dao := newTestService()

	started, svcErr := svc.CreatePlan(ctx, "gym", "alice", validPayload())
	require.Nil(t, svcErr)
	future := validPayload()
	future.StartsOn = day("2026-10-01")
	excused, svcErr := svc.CreatePlan(ctx, "gym", "alice", future)
	require.Nil(t, svcErr)
	dao.last[excused.ID] = day("2026-10-02")

	svc.clock = at("2026-09-22")
	svcErr = svc.DeletePlan(ctx, started.ID, "alice")
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusConflict, svcErr.Code, "already started")

	svcErr = svc.DeletePlan(ctx, excused.ID, "alice")
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusConflict, svcErr.Code, "has an excuse filed ahead")
}

func TestIsScheduledOn(t *testing.T) {
	endsOn := day("2026-09-30")
	plan := Plan{Monday: true, Friday: true, StartsOn: day("2026-09-21"), EndsOn: &endsOn}

	assert.True(t, plan.IsScheduledOn(day("2026-09-21")), "first Monday")
	assert.True(t, plan.IsScheduledOn(day("2026-09-25")), "Friday")
	assert.False(t, plan.IsScheduledOn(day("2026-09-23")), "Wednesday is not planned")
	assert.False(t, plan.IsScheduledOn(day("2026-09-14")), "before start")
	assert.False(t, plan.IsScheduledOn(day("2026-10-02")), "after end")
}
