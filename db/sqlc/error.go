package db

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

var (
	// Returned when a query expected a row but found none.
	ErrNoRows = pgx.ErrNoRows

	// Returned when too many rows are returned for a single-row query.
	ErrTooManyRows = errors.New("query returned too many rows")

	// Returned when the connection to the database is closed or lost.
	ErrConnClosed = errors.New("database connection closed")

	// Returned on serialization or deadlock failure (SQLSTATE 40001 or 40P01).
	ErrSerializationFailure = errors.New("serialization failure (retryable)")

	// Returned when an integrity constraint is violated (e.g. FK, unique).
	ErrIntegrityViolation = errors.New("integrity constraint violation")

	// Returned on unique constraint violation (SQLSTATE 23505).
	ErrUniqueViolation = pgconn.PgError{
		Code: "23505",
	}

	// Returned on foreign key violation (SQLSTATE 23503).
	ErrForeignKeyViolation = pgconn.PgError{
		Code: "23503",
	}

	// Returned on check constraint violation (SQLSTATE 23514).
	ErrCheckViolation = pgconn.PgError{
		Code: "23514",
	}

	// Returned on not null constraint violation (SQLSTATE 23502).
	ErrNotNullViolation = pgconn.PgError{
		Code: "23502",
	}
)
