package activities

import "time"

type Activity struct {
	ID          string    `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedBy   string    `db:"created_by"  json:"createdBy"`
	CreatedAt   time.Time `db:"created_at"  json:"createdAt"`
}

type CreateActivityPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
