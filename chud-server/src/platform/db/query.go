package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

func GetOne[T any](ctx context.Context, q Querier, query string, args ...any) (*T, error) {
	var result T
	err := sqlx.GetContext(ctx, q, &result, query, args...)
	switch {
	case err == nil:
		return &result, nil
	case errors.Is(err, sql.ErrNoRows), hasPgCode(err, pgInvalidTextInput), hasPgCode(err, pgForeignKeyViolation):
		return nil, ErrNotFound
	case IsUniqueViolation(err):
		return nil, ErrAlreadyExists
	default:
		return nil, err
	}
}

func WrapExec(ctx context.Context, operationName string, q Querier, query string, args ...any) error {
	_, err := Wrap(ctx, operationName, func() (struct{}, error) {
		result, err := q.ExecContext(ctx, query, args...)
		if hasPgCode(err, pgInvalidTextInput) {
			return struct{}{}, ErrNotFound
		}
		if hasPgCode(err, pgForeignKeyViolation) {
			return struct{}{}, ErrInUse
		}
		if err != nil {
			return struct{}{}, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return struct{}{}, err
		}
		if rows == 0 {
			return struct{}{}, ErrNotFound
		}
		return struct{}{}, nil
	})
	return err
}
