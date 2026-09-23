package users

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/platform/db"
)

type IUserDAO interface {
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	InsertUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	DemoteAdminsExcept(ctx context.Context, username string) error
}

type UserDAO struct {
	pool db.Querier
}

func NewUserDAO(pool db.Querier) IUserDAO {
	return &UserDAO{
		pool: pool,
	}
}

func (r *UserDAO) GetAllUsers(ctx context.Context) ([]User, error) {
	return db.Wrap(ctx, "GetAllUsers", func() ([]User, error) {
		users := []User{}
		err := sqlx.SelectContext(ctx, r.pool, &users, `SELECT * FROM users ORDER BY username`)
		return users, err
	})
}

func (r *UserDAO) GetUserByID(ctx context.Context, id string) (*User, error) {
	return db.Wrap(ctx, "GetUserByID", func() (*User, error) {
		return db.GetOne[User](ctx, r.pool, `SELECT * FROM users WHERE id = $1`, id)
	})
}

func (r *UserDAO) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return db.Wrap(ctx, "GetUserByUsername", func() (*User, error) {
		return db.GetOne[User](ctx, r.pool, `SELECT * FROM users WHERE username = $1`, username)
	})
}

func (r *UserDAO) InsertUser(ctx context.Context, user *User) (*User, error) {
	return db.Wrap(ctx, "InsertUser", func() (*User, error) {
		return db.GetOne[User](
			ctx,
			r.pool,
			`INSERT INTO users (id, username, password_hash, is_admin, color)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING *`,
			user.ID,
			user.Username,
			user.PasswordHash,
			user.IsAdmin,
			user.Color,
		)
	})
}

func (r *UserDAO) UpdateUser(ctx context.Context, user *User) (*User, error) {
	return db.Wrap(ctx, "UpdateUser", func() (*User, error) {
		return db.GetOne[User](
			ctx,
			r.pool,
			`UPDATE users
			SET username = $2, password_hash = $3, is_admin = $4, color = $5, updated_at = NOW()
			WHERE id = $1
			RETURNING *`,
			user.ID,
			user.Username,
			user.PasswordHash,
			user.IsAdmin,
			user.Color,
		)
	})
}

func (r *UserDAO) DeleteUser(ctx context.Context, id string) error {
	return db.WrapExec(ctx, "DeleteUser", r.pool, `DELETE FROM users WHERE id = $1`, id)
}

func (r *UserDAO) DemoteAdminsExcept(ctx context.Context, username string) error {
	_, err := db.Wrap(ctx, "DemoteAdminsExcept", func() (struct{}, error) {
		_, err := r.pool.ExecContext(
			ctx,
			`UPDATE users SET is_admin = FALSE, updated_at = NOW() WHERE is_admin AND username <> $1`,
			username,
		)
		return struct{}{}, err
	})
	return err
}
