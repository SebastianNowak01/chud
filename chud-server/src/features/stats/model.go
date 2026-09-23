package stats

import "github.com/sebnow/chud/platform/clock"

type Status string

const (
	StatusDone    Status = "done"
	StatusExcused Status = "excused"
	StatusMissed  Status = "missed"
	StatusPending Status = "pending"
)

const (
	PointsDone    = 3
	PointsExtra   = 1
	PointsExcused = 0
	PointsMissed  = -2
)

type Occurrence struct {
	Date       clock.Date `json:"date"`
	PlanID     string     `json:"planId"`
	PlanTitle  string     `json:"planTitle"`
	UserID     string     `json:"userId"`
	ActivityID string     `json:"activityId"`
	Status     Status     `json:"status"`
	EntryID    *string    `json:"entryId"`
}

type PlannedEntry struct {
	ID           string     `db:"id"`
	PlanID       string     `db:"plan_id"`
	ScheduledFor clock.Date `db:"scheduled_for"`
	Excused      bool       `db:"excused"`
}

type UnplannedCount struct {
	UserID string `db:"user_id"`
	Count  int    `db:"count"`
}

type UserStats struct {
	UserID  string   `json:"userId"`
	Points  int      `json:"points"`
	Done    int      `json:"done"`
	Extra   int      `json:"extra"`
	Excused int      `json:"excused"`
	Missed  int      `json:"missed"`
	Pending int      `json:"pending"`
	Rate    *float64 `json:"rate"`
}
