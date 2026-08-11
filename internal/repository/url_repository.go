package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohdrashid9678/xlink/internal/models"
)

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) (*models.URL, error)
	GetByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error)
	GetPublicByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	Update(ctx context.Context, userID uuid.UUID, shortCode string, update models.UpdateURLRequest) (*models.URL, error)
	Delete(ctx context.Context, userID uuid.UUID, shortCode string) error
	IncrementClickCount(ctx context.Context, shortCode string) error
}

type PostgresURLRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresURLRepository(pool *pgxpool.Pool) *PostgresURLRepository {
	return &PostgresURLRepository{pool: pool}
}

func (r *PostgresURLRepository) Create(ctx context.Context, url *models.URL) (*models.URL, error) {
	query := `INSERT INTO urls (id, user_id, short_code, long_url, custom_alias, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING ` + urlColumns
	created, err := scanURL(r.pool.QueryRow(ctx, query,
		url.ID, url.UserID, url.ShortCode, url.LongURL, url.CustomAlias, url.ExpiresAt,
	))
	return created, mapDatabaseError(err)
}

func (r *PostgresURLRepository) GetByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	url, err := scanURL(r.pool.QueryRow(ctx,
		`SELECT `+urlColumns+` FROM urls WHERE user_id = $1 AND short_code = $2`, userID, shortCode))
	return url, mapDatabaseError(err)
}

func (r *PostgresURLRepository) GetPublicByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	url, err := scanURL(r.pool.QueryRow(ctx, `SELECT `+urlColumns+` FROM urls WHERE short_code = $1
        AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`, shortCode))
	return url, mapDatabaseError(err)
}

func (r *PostgresURLRepository) Update(ctx context.Context, userID uuid.UUID, shortCode string, update models.UpdateURLRequest) (*models.URL, error) {
	sets := make([]string, 0, 4)
	args := []any{userID, shortCode}

	if update.LongURL != nil {
		args = append(args, *update.LongURL)
		sets = append(sets, fmt.Sprintf("long_url = $%d", len(args)))
	}
	if update.CustomAlias != nil {
		args = append(args, *update.CustomAlias)
		placeholder := fmt.Sprintf("$%d", len(args))
		// A custom alias is the public short code, so both values change together.
		sets = append(sets, "short_code = "+placeholder, "custom_alias = "+placeholder)
	}
	if update.ExpiresAt != nil {
		args = append(args, *update.ExpiresAt)
		sets = append(sets, fmt.Sprintf("expires_at = $%d", len(args)))
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	query := `UPDATE urls SET ` + strings.Join(sets, ", ") +
		` WHERE user_id = $1 AND short_code = $2 RETURNING ` + urlColumns
	url, err := scanURL(r.pool.QueryRow(ctx, query, args...))
	return url, mapDatabaseError(err)
}

func (r *PostgresURLRepository) Delete(ctx context.Context, userID uuid.UUID, shortCode string) error {
	commandTag, err := r.pool.Exec(ctx, `DELETE FROM urls WHERE user_id = $1 AND short_code = $2`, userID, shortCode)
	if err != nil {
		return mapDatabaseError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	commandTag, err := r.pool.Exec(ctx, `UPDATE urls
        SET click_count = click_count + 1, updated_at = CURRENT_TIMESTAMP
        WHERE short_code = $1`, shortCode)
	if err != nil {
		return mapDatabaseError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
