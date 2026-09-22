package plans

import "time"

type Plan struct {
	ID         string     `db:"id"          json:"id"`
	ActivityID string     `db:"activity_id" json:"activityId"`
	UserID     string     `db:"user_id"     json:"userId"`
	Title      string     `db:"title"       json:"title"`
	Monday     bool       `db:"monday"      json:"monday"`
	Tuesday    bool       `db:"tuesday"     json:"tuesday"`
	Wednesday  bool       `db:"wednesday"   json:"wednesday"`
	Thursday   bool       `db:"thursday"    json:"thursday"`
	Friday     bool       `db:"friday"      json:"friday"`
	Saturday   bool       `db:"saturday"    json:"saturday"`
	Sunday     bool       `db:"sunday"      json:"sunday"`
	StartsOn   time.Time  `db:"starts_on"   json:"startsOn"`
	EndsOn     *time.Time `db:"ends_on"     json:"endsOn"`
	CreatedAt  time.Time  `db:"created_at"  json:"createdAt"`
}

// PlanPayload creates or replaces a plan. Dates are YYYY-MM-DD.
type PlanPayload struct {
	Title     string  `json:"title"`
	Monday    bool    `json:"monday"`
	Tuesday   bool    `json:"tuesday"`
	Wednesday bool    `json:"wednesday"`
	Thursday  bool    `json:"thursday"`
	Friday    bool    `json:"friday"`
	Saturday  bool    `json:"saturday"`
	Sunday    bool    `json:"sunday"`
	StartsOn  string  `json:"startsOn"`
	EndsOn    *string `json:"endsOn"`
}

// IsScheduledOn reports whether the plan has a planned day on the given date.
func (p *Plan) IsScheduledOn(date time.Time) bool {
	if date.Before(p.StartsOn) || (p.EndsOn != nil && date.After(*p.EndsOn)) {
		return false
	}
	switch date.Weekday() {
	case time.Monday:
		return p.Monday
	case time.Tuesday:
		return p.Tuesday
	case time.Wednesday:
		return p.Wednesday
	case time.Thursday:
		return p.Thursday
	case time.Friday:
		return p.Friday
	case time.Saturday:
		return p.Saturday
	default:
		return p.Sunday
	}
}
