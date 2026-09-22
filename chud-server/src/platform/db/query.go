package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

// GetOne runs a query returning a single row scanned into T, mapping driver errors to sentinel errors.
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

// WrapExec runs a statement that must affect at least one row, returning ErrNotFound otherwise.
func WrapExec(ctx context.Context, operationName string, q Querier, query string, args ...any) error {
	_, err := Wrap(ctx, operationName, func() (struct{}, error) {
		result, err := q.ExecContext(ctx, query, args...)
		if hasPgCode(err, pgInvalidTextInput) {
			return struct{}{}, ErrNotFound
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
