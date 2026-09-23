package entries

import (
	"time"

	"github.com/sebnow/chud/platform/clock"
)

type Entry struct {
	ID           string      `db:"id"            json:"id"`
	ActivityID   string      `db:"activity_id"   json:"activityId"`
	UserID       string      `db:"user_id"       json:"userId"`
	PlanID       *string     `db:"plan_id"       json:"planId"`
	ScheduledFor *clock.Date `db:"scheduled_for" json:"scheduledFor"`
	Excused      bool        `db:"excused"       json:"excused"`
	Description  string      `db:"description"   json:"description"`
	OccurredAt   time.Time   `db:"occurred_at"   json:"occurredAt"`
	CreatedAt    time.Time   `db:"created_at"    json:"createdAt"`
}

type Media struct {
	ID          string `db:"id"           json:"id"`
	EntryID     string `db:"entry_id"     json:"entryId"`
	ContentType string `db:"content_type" json:"contentType"`
	Data        []byte `db:"data"         json:"-"`
}

type CreateEntryPayload struct {
	Description  string
	OccurredAt   time.Time
	PlanID       *string
	ScheduledFor *clock.Date
	Excused      bool
	Files        []Media
}
