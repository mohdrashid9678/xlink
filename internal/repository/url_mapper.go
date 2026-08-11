package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/mohdrashid9678/xlink/internal/models"
)

const urlColumns = `
    id, user_id, short_code, long_url, custom_alias, click_count,
    created_at, updated_at, expires_at`

func scanURL(row pgx.Row) (*models.URL, error) {
	url := new(models.URL)
	err := row.Scan(
		&url.ID,
		&url.UserID,
		&url.ShortCode,
		&url.LongURL,
		&url.CustomAlias,
		&url.ClickCount,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return url, nil
}
