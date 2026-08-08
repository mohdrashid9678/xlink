package models

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID `json:"id"`
	ShortCode   string    `json:"short_code"`
	LongUrl     string    `json:"long_url"`
	CustomAlias string    `json:"custom_alias"`
	ClickCount  uint      `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}
