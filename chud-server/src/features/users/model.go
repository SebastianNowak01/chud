package users

import "time"

type User struct {
	ID           string    `db:"id"            json:"id"`
	Username     string    `db:"username"      json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	IsAdmin      bool      `db:"is_admin"      json:"isAdmin"`
	Color        string    `db:"color"         json:"color"`
	CreatedAt    time.Time `db:"created_at"    json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updatedAt"`
}

type CreateUserPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateUserPayload updates the given fields; an empty password leaves it unchanged.
type UpdateUserPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateMePayload struct {
	Color string `json:"color"`
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
