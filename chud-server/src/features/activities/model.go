package activities

import "time"

type Activity struct {
	ID          string     `db:"id"          json:"id"`
	Name        string     `db:"name"        json:"name"`
	Description string     `db:"description" json:"description"`
	CreatedBy   string     `db:"created_by"  json:"createdBy"`
	CreatedAt   time.Time  `db:"created_at"  json:"createdAt"`
	ArchivedAt  *time.Time `db:"archived_at" json:"archivedAt"`
	HasHistory  bool       `db:"has_history" json:"hasHistory"`
}

func (a *Activity) IsArchived() bool {
	return a.ArchivedAt != nil
}

type CreateActivityPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
