package stats

import (
	"testing"

	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func day(s string) clock.Date {
	d, err := clock.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func statuses(occurrences []Occurrence) map[string]Status {
	result := map[string]Status{}
	for _, o := range occurrences {
		result[o.PlanID+" "+o.Date.String()] = o.Status
	}
	return result
}

func TestOccurrences(t *testing.T) {
	endsOn := day("2026-09-30")
	gym := plans.Plan{ID: "gym", UserID: "alice", ActivityID: "a-gym", Monday: true, Wednesday: true, StartsOn: day("2026-09-16"), EndsOn: &endsOn}
	run := plans.Plan{ID: "run", UserID: "alice", ActivityID: "a-gym", Wednesday: true, StartsOn: day("2026-09-01")}
	entries := []PlannedEntry{
		{ID: "e1", PlanID: "gym", ScheduledFor: day("2026-09-16")},
		{ID: "e2", PlanID: "gym", ScheduledFor: day("2026-09-21"), Excused: true},
		{ID: "e3", PlanID: "run", ScheduledFor: day("2026-09-30"), Excused: true},
	}

	got := Occurrences([]plans.Plan{gym, run}, entries, day("2026-09-14"), day("2026-10-05"), day("2026-09-23"))

	assert.Equal(t, map[string]Status{
		"gym 2026-09-16": StatusDone,
		"gym 2026-09-21": StatusExcused,
		"gym 2026-09-23": StatusPending,
		"gym 2026-09-28": StatusPending,
		"gym 2026-09-30": StatusPending,
		"run 2026-09-16": StatusMissed,
		"run 2026-09-23": StatusPending,
		"run 2026-09-30": StatusExcused,
	}, statuses(got), "starts on the 16th, ends on the 30th, overlapping plans count separately, today is pending")

	require.NotEmpty(t, got)
	assert.Equal(t, day("2026-09-16"), got[0].Date, "sorted by date")
	assert.Equal(t, "e1", *got[0].EntryID)
	assert.Equal(t, "alice", got[0].UserID)
}

func TestLeaderboard(t *testing.T) {
	userList := []users.User{{ID: "a", Username: "alice"}, {ID: "b", Username: "bob"}, {ID: "c", Username: "cyd"}, {ID: "d", Username: "dee"}}
	occurrences := []Occurrence{
		{UserID: "a", Status: StatusDone},
		{UserID: "a", Status: StatusDone},
		{UserID: "a", Status: StatusMissed},
		{UserID: "a", Status: StatusExcused},
		{UserID: "a", Status: StatusPending},
		{UserID: "b", Status: StatusDone},
		{UserID: "b", Status: StatusExcused},
		{UserID: "ghost", Status: StatusDone},
	}
	unplanned := []UnplannedCount{{UserID: "b", Count: 1}, {UserID: "c", Count: 4}}

	got := Leaderboard(userList, occurrences, unplanned)

	require.Len(t, got, 4)
	order := []string{got[0].UserID, got[1].UserID, got[2].UserID, got[3].UserID}
	assert.Equal(t, []string{"a", "b", "c", "d"}, order, "alice, bob and cyd tie at 4 points, more done ranks higher")

	alice := got[0]
	require.NotNil(t, alice.Rate)
	assert.InDelta(t, 2.0/3.0, *alice.Rate, 1e-9, "excused days do not count against the rate")
	alice.Rate = nil
	assert.Equal(t, UserStats{UserID: "a", Points: 4, Done: 2, Missed: 1, Excused: 1, Pending: 1}, alice)

	assert.Equal(t, UserStats{UserID: "c", Points: 4, Extra: 4}, got[2], "unplanned entries only, no rate")
	assert.Equal(t, UserStats{UserID: "d"}, got[3], "users without activity are listed too")
}
