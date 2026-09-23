package stats

import (
	"sort"

	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
)

type planDay struct {
	planID string
	date   clock.Date
}

func Occurrences(planList []plans.Plan, entries []PlannedEntry, from, to, today clock.Date) []Occurrence {
	resolved := make(map[planDay]PlannedEntry, len(entries))
	for _, e := range entries {
		resolved[planDay{e.PlanID, e.ScheduledFor}] = e
	}

	result := []Occurrence{}
	for _, plan := range planList {
		first, last := from, to
		if plan.StartsOn.After(first) {
			first = plan.StartsOn
		}
		if plan.EndsOn != nil && plan.EndsOn.Before(last) {
			last = *plan.EndsOn
		}

		for day := first; !day.After(last); day = day.AddDays(1) {
			if !plan.IsScheduledOn(day) {
				continue
			}
			occurrence := Occurrence{
				Date:       day,
				PlanID:     plan.ID,
				PlanTitle:  plan.Title,
				UserID:     plan.UserID,
				ActivityID: plan.ActivityID,
				Status:     StatusPending,
			}
			if entry, ok := resolved[planDay{plan.ID, day}]; ok {
				occurrence.Status = StatusDone
				if entry.Excused {
					occurrence.Status = StatusExcused
				}
				occurrence.EntryID = &entry.ID
			} else if day.Before(today) {
				occurrence.Status = StatusMissed
			}
			result = append(result, occurrence)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Date != result[j].Date {
			return result[i].Date.Before(result[j].Date)
		}
		return result[i].PlanID < result[j].PlanID
	})
	return result
}

func Leaderboard(userList []users.User, occurrences []Occurrence, unplanned []UnplannedCount) []UserStats {
	byUser := make(map[string]*UserStats, len(userList))
	names := make(map[string]string, len(userList))
	result := make([]*UserStats, 0, len(userList))
	for _, u := range userList {
		stats := &UserStats{UserID: u.ID}
		byUser[u.ID] = stats
		names[u.ID] = u.Username
		result = append(result, stats)
	}

	for _, o := range occurrences {
		stats, ok := byUser[o.UserID]
		if !ok {
			continue
		}
		switch o.Status {
		case StatusDone:
			stats.Done++
		case StatusExcused:
			stats.Excused++
		case StatusMissed:
			stats.Missed++
		case StatusPending:
			stats.Pending++
		}
	}
	for _, c := range unplanned {
		if stats, ok := byUser[c.UserID]; ok {
			stats.Extra = c.Count
		}
	}

	for _, s := range result {
		s.Points = s.Done*PointsDone + s.Extra*PointsExtra + s.Excused*PointsExcused + s.Missed*PointsMissed
		if judged := s.Done + s.Missed; judged > 0 {
			rate := float64(s.Done) / float64(judged)
			s.Rate = &rate
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.Done != b.Done {
			return a.Done > b.Done
		}
		return names[a.UserID] < names[b.UserID]
	})

	out := make([]UserStats, len(result))
	for i, s := range result {
		out[i] = *s
	}
	return out
}
