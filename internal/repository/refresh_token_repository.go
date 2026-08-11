package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohdrashid9678/xlink/internal/models"
)

var ErrInvalidRefreshToken = pgx.ErrNoRows

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	Rotate(ctx context.Context, currentHash string, replacement *models.RefreshToken) (uuid.UUID, error)
	Revoke(ctx context.Context, tokenHash string) error
}

type PostgresRefreshTokenRepository struct{ pool *pgxpool.Pool }

func NewPostgresRefreshTokenRepository(pool *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{pool: pool}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO refresh_tokens (token, user_id, expires_at)
        VALUES ($1, $2, $3)`, token.TokenHash, token.UserID, token.ExpiresAt)
	return mapDatabaseError(err)
}

func (r *PostgresRefreshTokenRepository) Rotate(ctx context.Context, currentHash string, replacement *models.RefreshToken) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	var expiresAt time.Time
	var revokedAt *time.Time
	err = tx.QueryRow(ctx, `SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1 FOR UPDATE`, currentHash).
		Scan(&userID, &expiresAt, &revokedAt)
	if err != nil || revokedAt != nil || !expiresAt.After(time.Now().UTC()) {
		return uuid.Nil, ErrInvalidRefreshToken
	}
	if _, err = tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE token = $1`, currentHash); err != nil {
		return uuid.Nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`, replacement.TokenHash, userID, replacement.ExpiresAt); err != nil {
		return uuid.Nil, mapDatabaseError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
        WHERE token = $1 AND revoked_at IS NULL`, tokenHash)
	return mapDatabaseError(err)
}
