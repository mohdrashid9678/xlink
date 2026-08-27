package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohdrashid9678/xlink/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.Create",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()

	query := `INSERT INTO urls (id, user_id, short_code, long_url, custom_alias, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING ` + urlColumns
	created, err := scanURL(r.pool.QueryRow(ctx, query,
		url.ID, url.UserID, url.ShortCode, url.LongURL, url.CustomAlias, url.ExpiresAt,
	))
	return created, mapDatabaseError(err)
}

func (r *PostgresURLRepository) GetByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.GetByShortCode",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()

	url, err := scanURL(r.pool.QueryRow(ctx,
		`SELECT `+urlColumns+` FROM urls WHERE user_id = $1 AND short_code = $2`, userID, shortCode))
	return url, mapDatabaseError(err)
}

func (r *PostgresURLRepository) GetPublicByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.GetPublicByShortCode",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()

	url, err := scanURL(r.pool.QueryRow(ctx, `SELECT `+urlColumns+` FROM urls WHERE short_code = $1
        AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`, shortCode))
	return url, mapDatabaseError(err)
}

func (r *PostgresURLRepository) Update(ctx context.Context, userID uuid.UUID, shortCode string, update models.UpdateURLRequest) (*models.URL, error) {
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.Update",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()

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

	query := fmt.Sprintf(`UPDATE urls SET %s, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND short_code = $2 RETURNING %s`,
		strings.Join(sets, ", "), urlColumns)
	updated, err := scanURL(r.pool.QueryRow(ctx, query, args...))
	return updated, mapDatabaseError(err)
}

func (r *PostgresURLRepository) Delete(ctx context.Context, userID uuid.UUID, shortCode string) error {
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.Delete",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()

	cmd, err := r.pool.Exec(ctx, `DELETE FROM urls WHERE user_id = $1 AND short_code = $2`, userID, shortCode)
	if err != nil {
		return mapDatabaseError(err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.IncrementClickCount",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()

	cmd, err := r.pool.Exec(ctx, `UPDATE urls SET click_count = click_count + 1 WHERE short_code = $1`, shortCode)
	if err != nil {
		return mapDatabaseError(err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
