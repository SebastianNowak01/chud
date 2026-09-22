package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	config "github.com/sebnow/chud/platform/config"
	"github.com/sebnow/chud/platform/log"

	// Blank import to register the "pgx" database/sql driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed schema.sql
var dbSchema string

var ErrNotFound = errors.New("resource not found")
var ErrAlreadyExists = errors.New("resource already exists")

const pgUniqueViolation = "23505"

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

const (
	connectTimeout  = 30 * time.Second
	connectInterval = 1 * time.Second
)

// Querier is the query surface shared by *sqlx.DB and *sqlx.Tx.
type Querier = sqlx.ExtContext

// Client owns the PostgreSQL connection pool.
type Client struct {
	sqlxDB *sqlx.DB
}

// New creates a new DB client using the provided DSN, retrying until the database accepts connections.
func New(ctx context.Context, dsn string) (*Client, error) {
	logger := log.FromContext(ctx)

	pool, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	for {
		err = pool.PingContext(ctx)
		if err == nil {
			break
		}
		logger.Warn().Err(err).Msg("Database not ready, retrying...")
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		case <-time.After(connectInterval):
		}
	}

	c := &Client{sqlxDB: pool}

	if err := c.initSchema(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return c, nil
}

// NewFromEnv creates a new DB client using the DSN from environment.
func NewFromEnv(ctx context.Context) (*Client, error) {
	log.FromContext(ctx).Info().Msg("Connecting to PostgreSQL...")
	return New(ctx, os.Getenv(config.DatabaseURL))
}

// Querier returns the connection pool as a Querier.
func (c *Client) Querier() Querier {
	return c.sqlxDB
}

// initSchema executes the embedded schema to set up the database structure.
func (c *Client) initSchema(ctx context.Context) error {
	if _, err := c.sqlxDB.ExecContext(ctx, dbSchema); err != nil {
		return err
	}

	log.FromContext(ctx).Info().Msg("Database schema initialized successfully")
	return nil
}

// Close closes the underlying connection pool.
func (c *Client) Close() {
	if c.sqlxDB != nil {
		c.sqlxDB.Close()
	}
}

// Wrap runs a DB operation, logging its duration and any unexpected error.
func Wrap[T any](ctx context.Context, operationName string, operation func() (T, error)) (T, error) {
	logger := log.FromContext(ctx)

	start := time.Now()
	result, err := operation()
	elapsed := time.Since(start)

	switch {
	case err == nil:
		logger.Trace().Dur("duration_ms", elapsed).Msgf("DB operation %s completed", operationName)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrAlreadyExists):
		// Expected outcomes, handled by the caller.
	default:
		logger.Error().Err(err).Dur("duration_ms", elapsed).Msgf("DB operation %s failed", operationName)
	}

	return result, err
}
