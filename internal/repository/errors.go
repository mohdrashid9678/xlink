package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// These errors describe storage outcomes without leaking PostgreSQL details to
// the service layer.
var (
	ErrNotFound = errors.New("repository: record not found")
	ErrConflict = errors.New("repository: unique constraint violation")
)

func mapDatabaseError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}

	return err
}
