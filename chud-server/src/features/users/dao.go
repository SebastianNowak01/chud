package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sebnow/chud/platform/db"
)

const userColumns = `id, username, password_hash, is_admin, created_at, updated_at`

type IUserDAO interface {
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	InsertUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	// DemoteAdminsExcept clears the admin flag of every user other than username.
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
		err := sqlx.SelectContext(
			ctx,
			r.pool,
			&users,
			`SELECT `+userColumns+` FROM users ORDER BY created_at, username`,
		)
		return users, err
	})
}

func (r *UserDAO) GetUserByID(ctx context.Context, id string) (*User, error) {
	return db.Wrap(ctx, "GetUserByID", func() (*User, error) {
		return r.getOne(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	})
}

func (r *UserDAO) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return db.Wrap(ctx, "GetUserByUsername", func() (*User, error) {
		return r.getOne(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username)
	})
}

func (r *UserDAO) InsertUser(ctx context.Context, user *User) (*User, error) {
	return db.Wrap(ctx, "InsertUser", func() (*User, error) {
		return r.getOne(
			ctx,
			`INSERT INTO users (id, username, password_hash, is_admin)
			VALUES ($1, $2, $3, $4)
			RETURNING `+userColumns,
			user.ID,
			user.Username,
			user.PasswordHash,
			user.IsAdmin,
		)
	})
}

func (r *UserDAO) UpdateUser(ctx context.Context, user *User) (*User, error) {
	return db.Wrap(ctx, "UpdateUser", func() (*User, error) {
		return r.getOne(
			ctx,
			`UPDATE users
			SET username = $2, password_hash = $3, is_admin = $4, updated_at = NOW()
			WHERE id = $1
			RETURNING `+userColumns,
			user.ID,
			user.Username,
			user.PasswordHash,
			user.IsAdmin,
		)
	})
}

func (r *UserDAO) DeleteUser(ctx context.Context, id string) error {
	_, err := db.Wrap(ctx, "DeleteUser", func() (struct{}, error) {
		result, err := r.pool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
		if err != nil {
			return struct{}{}, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return struct{}{}, err
		}
		if rows == 0 {
			return struct{}{}, db.ErrNotFound
		}
		return struct{}{}, nil
	})
	return err
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

// getOne runs a query returning a single user row, mapping driver errors to db sentinel errors.
func (r *UserDAO) getOne(ctx context.Context, query string, args ...any) (*User, error) {
	var user User
	err := r.pool.QueryRowxContext(ctx, query, args...).StructScan(&user)
	switch {
	case err == nil:
		return &user, nil
	case errors.Is(err, sql.ErrNoRows):
		return nil, db.ErrNotFound
	case db.IsUniqueViolation(err):
		return nil, db.ErrAlreadyExists
	default:
		return nil, err
	}
}
