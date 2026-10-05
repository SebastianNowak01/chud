package summaries

import (
	"strings"
	"testing"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func d(value string) clock.Date {
	date, err := clock.ParseDate(value)
	if err != nil {
		panic(err)
	}
	return date
}

func ptr[T any](v T) *T { return &v }

var (
	monday       = d("2026-09-28")
	testUsers    = []users.User{{ID: "alice", Username: "Alice"}, {ID: "bob", Username: "Bob"}, {ID: "carol", Username: "Carol"}}
	testActivity = []activities.Activity{{ID: "gym", Name: "Siłownia"}, {ID: "run", Name: "Bieganie", ArchivedAt: ptr(time.Now())}}
)

func occurrence(date, user, activity string, status stats.Status) stats.Occurrence {
	return stats.Occurrence{Date: d(date), UserID: user, ActivityID: activity, Status: status}
}

func TestBuildFacts(t *testing.T) {
	facts := BuildFacts(FactsInput{
		Week:  monday,
		Today: d("2026-10-01"),
		Occurrences: []stats.Occurrence{
			occurrence("2026-09-28", "alice", "gym", stats.StatusDone),
			occurrence("2026-09-29", "alice", "gym", stats.StatusDone),
			occurrence("2026-09-28", "bob", "run", stats.StatusMissed),
			occurrence("2026-09-29", "bob", "run", stats.StatusExcused),
			occurrence("2026-10-02", "bob", "gym", stats.StatusPending),
		},
		Leaderboard: []stats.UserStats{
			{UserID: "alice", Points: 7, Done: 2, Extra: 1, Rate: ptr(1.0)},
			{UserID: "bob", Points: -2, Missed: 1, Excused: 1, Pending: 1, Rate: ptr(0.0)},
			{UserID: "carol"},
		},
		LastWeek: []stats.UserStats{
			{UserID: "bob", Points: 6, Done: 2, Rate: ptr(2.0 / 3)},
			{UserID: "alice", Points: 3, Done: 1},
			{UserID: "carol"},
		},
		Users:      testUsers,
		Activities: testActivity,
		Entries: []entries.Entry{
			{UserID: "alice", ActivityID: "gym", OccurredAt: d("2026-09-30").Start().Add(8 * time.Hour), Description: "  Nowy rekord  "},
			{UserID: "bob", ActivityID: "run", PlanID: ptr("p"), ScheduledFor: ptr(d("2026-09-29")), Excused: true,
				OccurredAt: d("2026-10-01").Start(), Description: strings.Repeat("a", maxNoteLength+10)},
			{UserID: "alice", ActivityID: "gym", OccurredAt: d("2026-10-05").Start()},
		},
	})

	assert.Equal(t, d("2026-10-04"), facts.WeekEnd)
	assert.False(t, facts.Finished)
	assert.Equal(t, Counts{Done: 2, Extra: 1, Excused: 1, Missed: 1, Pending: 1, RatePercent: ptr(67)}, facts.Team)

	require.Len(t, facts.People, 2, "people without activity in either week are skipped")
	alice := facts.People[0]
	assert.Equal(t, "Alice", alice.Name)
	assert.Equal(t, 1, alice.Rank)
	assert.True(t, alice.PerfectWeek)
	assert.Equal(t, &LastWeekFacts{Rank: 2, Points: 3}, alice.LastWeek)
	assert.Equal(t, ptr(100), alice.RatePercent)
	bob := facts.People[1]
	assert.False(t, bob.PerfectWeek)
	assert.Equal(t, &LastWeekFacts{Rank: 1, Points: 6, RatePercent: ptr(67)}, bob.LastWeek)

	require.Len(t, facts.Activities, 2)
	assert.Equal(t, ActivityFacts{Name: "Siłownia", Counts: Counts{Done: 2, Extra: 1, Pending: 1, RatePercent: ptr(100)}}, facts.Activities[0])
	assert.True(t, facts.Activities[1].Archived)

	require.Len(t, facts.Days, 4)
	assert.Equal(t, "poniedziałek", facts.Days[0].Weekday)
	assert.Equal(t, Counts{Extra: 1}, facts.Days[2].Counts)

	require.Len(t, facts.Excuses, 1)
	assert.Equal(t, d("2026-09-29"), facts.Excuses[0].Date, "excuse belongs to its scheduled day")
	assert.Equal(t, "Bob", facts.Excuses[0].Person)
	assert.Equal(t, maxNoteLength+1, len([]rune(facts.Excuses[0].Text)))
	assert.Equal(t, []NoteFacts{{Person: "Alice", Activity: "Siłownia", Date: d("2026-09-30"), Text: "Nowy rekord"}}, facts.Notes)
	assert.False(t, facts.Empty())
}

func TestBuildFactsEmptyWeek(t *testing.T) {
	facts := BuildFacts(FactsInput{Week: monday, Today: d("2026-10-10"), Leaderboard: []stats.UserStats{{UserID: "alice"}}, Users: testUsers})
	assert.True(t, facts.Empty())
	assert.True(t, facts.Finished)
	assert.Empty(t, facts.People)
}

func TestEncodeIsStable(t *testing.T) {
	in := FactsInput{Week: monday, Today: monday, Occurrences: []stats.Occurrence{occurrence("2026-09-28", "alice", "gym", stats.StatusDone)}}
	_, first := BuildFacts(in).Encode()
	_, second := BuildFacts(in).Encode()
	assert.Equal(t, first, second)

	in.Occurrences[0].Status = stats.StatusMissed
	_, changed := BuildFacts(in).Encode()
	assert.NotEqual(t, first, changed)
}

func TestWeekExpiry(t *testing.T) {
	sundayEvening := d("2026-10-04").Start().Add(20 * time.Hour)
	assert.Equal(t, ptr(d("2026-10-05").Start()), weekExpiry(monday, d("2026-10-04"), sundayEvening), "current week expires at its end")

	wednesday := d("2026-09-30").Start().Add(9 * time.Hour)
	assert.Equal(t, ptr(wednesday.Add(24*time.Hour)), weekExpiry(monday, d("2026-09-30"), wednesday))

	assert.Nil(t, weekExpiry(monday, d("2026-10-06"), d("2026-10-06").Start()), "past weeks never expire")
}

func TestCleanResponse(t *testing.T) {
	assert.Equal(t, "Dobry tydzień.", cleanResponse("<think>\nliczę\n</think>\n\n Dobry tydzień. "))
	assert.Equal(t, "Dobry tydzień. Ania prowadzi!", cleanResponse("Dobry tydzień. Ania prowadzi! A Bartek w tym tyg"), "unfinished sentence is cut")
	assert.Empty(t, cleanResponse("<think>liczę punkty i jeszcze"), "unclosed reasoning is dropped")
	assert.Equal(t, maxSummaryLength+1, len([]rune(cleanResponse(strings.Repeat("x", maxSummaryLength+5)))))
}
