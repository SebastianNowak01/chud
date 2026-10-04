package summaries

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
)

const (
	maxNotes      = 8
	maxNoteLength = 280
)

var weekdayNames = [...]string{"niedziela", "poniedziałek", "wtorek", "środa", "czwartek", "piątek", "sobota"}

type Counts struct {
	Done        int  `json:"done"`
	Extra       int  `json:"extra"`
	Excused     int  `json:"excused"`
	Missed      int  `json:"missed"`
	Pending     int  `json:"pending"`
	RatePercent *int `json:"ratePercent"`
}

type LastWeekFacts struct {
	Rank        int  `json:"rank"`
	Points      int  `json:"points"`
	RatePercent *int `json:"ratePercent"`
}

type PersonFacts struct {
	Name        string         `json:"name"`
	Rank        int            `json:"rank"`
	Points      int            `json:"points"`
	PerfectWeek bool           `json:"perfectWeek"`
	LastWeek    *LastWeekFacts `json:"lastWeek"`
	Counts
}

type ActivityFacts struct {
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
	Counts
}

type DayFacts struct {
	Date    clock.Date `json:"date"`
	Weekday string     `json:"weekday"`
	Counts
}

type NoteFacts struct {
	Person   string     `json:"person"`
	Activity string     `json:"activity"`
	Date     clock.Date `json:"date"`
	Text     string     `json:"text"`
}

type WeekFacts struct {
	WeekStart  clock.Date      `json:"weekStart"`
	WeekEnd    clock.Date      `json:"weekEnd"`
	Finished   bool            `json:"weekFinished"`
	Team       Counts          `json:"team"`
	People     []PersonFacts   `json:"people"`
	Activities []ActivityFacts `json:"activities"`
	Days       []DayFacts      `json:"days"`
	Excuses    []NoteFacts     `json:"excuses"`
	Notes      []NoteFacts     `json:"entryNotes"`
}

type FactsInput struct {
	Week        clock.Date
	Today       clock.Date
	Occurrences []stats.Occurrence
	Leaderboard []stats.UserStats
	LastWeek    []stats.UserStats
	Users       []users.User
	Activities  []activities.Activity
	Entries     []entries.Entry
}

func BuildFacts(in FactsInput) WeekFacts {
	end := in.Week.AddDays(6)
	facts := WeekFacts{
		WeekStart:  in.Week,
		WeekEnd:    end,
		Finished:   in.Today.After(end),
		People:     []PersonFacts{},
		Activities: []ActivityFacts{},
		Days:       []DayFacts{},
		Excuses:    []NoteFacts{},
		Notes:      []NoteFacts{},
	}

	userNames := make(map[string]string, len(in.Users))
	for _, u := range in.Users {
		userNames[u.ID] = u.Username
	}
	activityByID := make(map[string]activities.Activity, len(in.Activities))
	for _, a := range in.Activities {
		activityByID[a.ID] = a
	}

	byActivity := map[string]*Counts{}
	byDay := map[clock.Date]*Counts{}
	countFor := func(m map[string]*Counts, key string) *Counts {
		if m[key] == nil {
			m[key] = &Counts{}
		}
		return m[key]
	}
	dayFor := func(day clock.Date) *Counts {
		if byDay[day] == nil {
			byDay[day] = &Counts{}
		}
		return byDay[day]
	}

	for _, o := range in.Occurrences {
		for _, c := range []*Counts{&facts.Team, countFor(byActivity, o.ActivityID), dayFor(o.Date)} {
			addStatus(c, o.Status)
		}
	}

	for _, e := range in.Entries {
		day := clock.DateOf(e.OccurredAt)
		if e.ScheduledFor != nil {
			day = *e.ScheduledFor
		}
		if e.PlanID == nil {
			if day.Before(in.Week) || day.After(end) {
				continue
			}
			facts.Team.Extra++
			countFor(byActivity, e.ActivityID).Extra++
			dayFor(day).Extra++
		}
		text := truncate(strings.TrimSpace(e.Description))
		if text == "" {
			continue
		}
		note := NoteFacts{Person: userNames[e.UserID], Activity: activityByID[e.ActivityID].Name, Date: day, Text: text}
		if e.Excused && len(facts.Excuses) < maxNotes {
			facts.Excuses = append(facts.Excuses, note)
		} else if !e.Excused && len(facts.Notes) < maxNotes {
			facts.Notes = append(facts.Notes, note)
		}
	}
	setRate(&facts.Team)

	lastWeek := make(map[string]LastWeekFacts, len(in.LastWeek))
	for i, s := range in.LastWeek {
		if active(s) {
			lastWeek[s.UserID] = LastWeekFacts{Rank: i + 1, Points: s.Points, RatePercent: percent(s.Rate)}
		}
	}
	rank := 0
	for _, s := range in.Leaderboard {
		previous, hadLastWeek := lastWeek[s.UserID]
		if !active(s) && !hadLastWeek {
			continue
		}
		rank++
		person := PersonFacts{
			Name:   userNames[s.UserID],
			Rank:   rank,
			Points: s.Points,
			Counts: Counts{
				Done: s.Done, Extra: s.Extra, Excused: s.Excused, Missed: s.Missed, Pending: s.Pending,
				RatePercent: percent(s.Rate),
			},
			PerfectWeek: s.Done > 0 && s.Missed == 0 && s.Excused == 0 && s.Pending == 0,
		}
		if hadLastWeek {
			person.LastWeek = &previous
		}
		facts.People = append(facts.People, person)
	}

	for _, a := range in.Activities {
		c, ok := byActivity[a.ID]
		if !ok {
			continue
		}
		setRate(c)
		facts.Activities = append(facts.Activities, ActivityFacts{Name: a.Name, Archived: a.IsArchived(), Counts: *c})
	}

	for day := in.Week; !day.After(end); day = day.AddDays(1) {
		c, ok := byDay[day]
		if !ok {
			continue
		}
		setRate(c)
		facts.Days = append(facts.Days, DayFacts{Date: day, Weekday: weekdayNames[day.Weekday()], Counts: *c})
	}

	return facts
}

func (f WeekFacts) Empty() bool {
	t := f.Team
	return t.Done+t.Extra+t.Excused+t.Missed+t.Pending == 0 && len(f.Notes) == 0
}

func (f WeekFacts) Encode() (data string, hash string) {
	raw, _ := json.Marshal(f)
	sum := sha256.Sum256(raw)
	return string(raw), hex.EncodeToString(sum[:])
}

func addStatus(c *Counts, status stats.Status) {
	switch status {
	case stats.StatusDone:
		c.Done++
	case stats.StatusExcused:
		c.Excused++
	case stats.StatusMissed:
		c.Missed++
	case stats.StatusPending:
		c.Pending++
	}
}

func setRate(c *Counts) {
	if judged := c.Done + c.Missed; judged > 0 {
		rate := float64(c.Done) / float64(judged)
		c.RatePercent = percent(&rate)
	}
}

func percent(rate *float64) *int {
	if rate == nil {
		return nil
	}
	value := int(math.Round(*rate * 100))
	return &value
}

func active(s stats.UserStats) bool {
	return s.Done+s.Extra+s.Excused+s.Missed+s.Pending > 0
}

func truncate(text string) string {
	runes := []rune(text)
	if len(runes) <= maxNoteLength {
		return text
	}
	return strings.TrimSpace(string(runes[:maxNoteLength])) + "…"
}

func weekStart(day clock.Date) clock.Date {
	return day.AddDays(-((int(day.Weekday()) + 6) % 7))
}

func weekExpiry(week, today clock.Date, now time.Time) *time.Time {
	if week != weekStart(today) {
		return nil
	}
	expires := now.Add(currentWeekTTL)
	if end := week.AddDays(7).Start(); end.Before(expires) {
		expires = end
	}
	return &expires
}
