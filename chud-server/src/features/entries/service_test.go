package entries

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeActivityDAO struct{ activities.IActivityDAO }

func (fakeActivityDAO) GetActivityByID(_ context.Context, id string) (*activities.Activity, error) {
	if id != "gym" && id != "walk" {
		return nil, db.ErrNotFound
	}
	return &activities.Activity{ID: id}, nil
}

type fakePlanDAO struct{ plans.IPlanDAO }

func (fakePlanDAO) GetPlanByID(_ context.Context, id string) (*plans.Plan, error) {
	if id != "alice-gym" {
		return nil, db.ErrNotFound
	}
	return &plans.Plan{ID: id, ActivityID: "gym", UserID: "alice", Monday: true, StartsOn: *date("2026-09-21")}, nil
}

type fakeEntryDAO struct {
	IEntryDAO
	entries []Entry
	media   []Media
}

func (d *fakeEntryDAO) InsertEntry(_ context.Context, entry *Entry) (*Entry, error) {
	for _, e := range d.entries {
		if e.PlanID != nil && entry.PlanID != nil && *e.PlanID == *entry.PlanID &&
			*e.ScheduledFor == *entry.ScheduledFor {
			return nil, db.ErrAlreadyExists
		}
	}
	d.entries = append(d.entries, *entry)
	return entry, nil
}

func (d *fakeEntryDAO) InsertMedia(_ context.Context, _ string, media []Media) error {
	d.media = append(d.media, media...)
	return nil
}

type fakeDB struct{}

func (fakeDB) Querier() db.Querier { return nil }

func (fakeDB) WithTx(_ context.Context, fn func(q db.Querier) error) error { return fn(nil) }

func (d *fakeEntryDAO) GetEntriesInRange(_ context.Context, from, to time.Time) ([]Entry, error) {
	var result []Entry
	for _, e := range d.entries {
		if !e.OccurredAt.Before(from) && e.OccurredAt.Before(to) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (d *fakeEntryDAO) GetUserEntriesInRange(ctx context.Context, userID string, from, to time.Time) ([]Entry, error) {
	all, _ := d.GetEntriesInRange(ctx, from, to)
	var result []Entry
	for _, e := range all {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

func newTestService() (*EntryService, *fakeEntryDAO) {
	dao := &fakeEntryDAO{}
	return NewEntryService(EntryServiceDeps{
		DB:          fakeDB{},
		NewEntryDAO: func(db.Querier) IEntryDAO { return dao },
		ActivityDAO: fakeActivityDAO{},
		PlanDAO:     fakePlanDAO{},
		Clock:       clock.Fixed(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)),
	}), dao
}

func ptr[T any](v T) *T { return &v }

func date(s string) *clock.Date {
	d, err := clock.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return &d
}

func TestCreateUnplannedEntry(t *testing.T) {
	svc, dao := newTestService()

	entry, svcErr := svc.CreateEntry(context.Background(), "gym", "bob", CreateEntryPayload{
		Description: " leg day ",
		Files:       []Media{{ContentType: "image/png", Data: []byte("png")}},
	})
	require.Nil(t, svcErr)
	assert.Equal(t, "leg day", entry.Description)
	assert.False(t, entry.OccurredAt.IsZero(), "defaults to now")
	assert.Len(t, dao.media, 1)
}

func TestCreatePlannedEntries(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestService()

	_, svcErr := svc.CreateEntry(ctx, "gym", "alice", CreateEntryPayload{
		PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-21"), OccurredAt: time.Date(2026, 9, 21, 18, 0, 0, 0, time.UTC),
	})
	require.Nil(t, svcErr)

	_, svcErr = svc.CreateEntry(ctx, "gym", "alice", CreateEntryPayload{
		PlanID: ptr("alice-gym"), ScheduledFor: date("2026-10-12"), Excused: true, Description: "trip",
	})
	require.Nil(t, svcErr, "excuses can be filed ahead")

	_, svcErr = svc.CreateEntry(ctx, "gym", "alice", CreateEntryPayload{
		PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-28"), Excused: true, Description: "sick",
	})
	require.Nil(t, svcErr)

	_, svcErr = svc.CreateEntry(ctx, "gym", "alice", CreateEntryPayload{
		PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-21"), OccurredAt: time.Date(2026, 9, 21, 19, 0, 0, 0, time.UTC),
	})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusConflict, svcErr.Code, "day already resolved")
}

func TestCreateEntryValidation(t *testing.T) {
	tests := []struct {
		name     string
		activity string
		user     string
		payload  CreateEntryPayload
		wantCode int
	}{
		{"unknown activity", "nope", "bob", CreateEntryPayload{}, http.StatusNotFound},
		{"excuse without plan", "gym", "bob", CreateEntryPayload{Excused: true, Description: "x"}, http.StatusBadRequest},
		{"day without plan", "gym", "bob", CreateEntryPayload{ScheduledFor: date("2026-09-21")}, http.StatusBadRequest},
		{"unknown plan", "gym", "alice", CreateEntryPayload{PlanID: ptr("nope"), ScheduledFor: date("2026-09-21")}, http.StatusNotFound},
		{"someone else's plan", "gym", "bob", CreateEntryPayload{PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-21")}, http.StatusForbidden},
		{"plan of other activity", "walk", "alice", CreateEntryPayload{PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-21")}, http.StatusBadRequest},
		{"plan without day", "gym", "alice", CreateEntryPayload{PlanID: ptr("alice-gym")}, http.StatusBadRequest},
		{"day not planned", "gym", "alice", CreateEntryPayload{PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-22")}, http.StatusBadRequest},
		{"planned day done on another day", "gym", "alice", CreateEntryPayload{PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-21"), OccurredAt: time.Date(2026, 9, 25, 9, 0, 0, 0, time.UTC)}, http.StatusBadRequest},
		{"future planned day done now", "gym", "alice", CreateEntryPayload{PlanID: ptr("alice-gym"), ScheduledFor: date("2026-10-12")}, http.StatusBadRequest},
		{"entry in the future", "gym", "bob", CreateEntryPayload{OccurredAt: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)}, http.StatusBadRequest},
		{"excuse without description", "gym", "alice", CreateEntryPayload{PlanID: ptr("alice-gym"), ScheduledFor: date("2026-09-21"), Excused: true}, http.StatusBadRequest},
		{"not a media file", "gym", "bob", CreateEntryPayload{Files: []Media{{ContentType: "application/pdf"}}}, http.StatusBadRequest},
		{"file too big", "gym", "bob", CreateEntryPayload{Files: []Media{{ContentType: "image/png", Data: make([]byte, MaxFileSize+1)}}}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService()
			_, svcErr := svc.CreateEntry(context.Background(), tt.activity, tt.user, tt.payload)
			require.NotNil(t, svcErr)
			assert.Equal(t, tt.wantCode, svcErr.Code)
		})
	}
}

func TestGetEntriesInRange(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestService()
	at := func(s string) time.Time {
		ts, err := time.Parse(time.RFC3339, s)
		require.NoError(t, err)
		return ts
	}

	for _, e := range []struct{ user, at string }{
		{"alice", "2026-09-01T00:00:00Z"},
		{"bob", "2026-09-15T12:00:00Z"},
		{"alice", "2026-09-30T23:59:59Z"},
		{"alice", "2026-10-01T00:00:00Z"},
	} {
		_, svcErr := svc.CreateEntry(ctx, "gym", e.user, CreateEntryPayload{OccurredAt: at(e.at)})
		require.Nil(t, svcErr)
	}

	from, to := *date("2026-09-01"), *date("2026-09-30")

	all, svcErr := svc.GetEntriesInRange(ctx, from, to)
	require.Nil(t, svcErr)
	assert.Len(t, all, 3, "both days are included, the day after is not")

	mine, svcErr := svc.GetUserEntriesInRange(ctx, "alice", from, to)
	require.Nil(t, svcErr)
	require.Len(t, mine, 2)
	assert.Equal(t, "alice", mine[0].UserID)
}

func TestGetEntriesInRangeValidation(t *testing.T) {
	svc, _ := newTestService()
	today := *date("2026-10-05")

	_, svcErr := svc.GetEntriesInRange(context.Background(), today, today.AddDays(-1))
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusBadRequest, svcErr.Code, "reversed range")

	_, svcErr = svc.GetEntriesInRange(context.Background(), today.AddDays(-2*365), today)
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusBadRequest, svcErr.Code, "range too long")
}
