package models

import "time"

// CreateURLRequest is the accepted payload for creating a short URL.
type CreateURLRequest struct {
	LongURL     string     `json:"long_url"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UpdateURLRequest deliberately contains only fields clients are allowed to change.
// A non-nil field is applied; omitted fields remain unchanged.
type UpdateURLRequest struct {
	LongURL     *string    `json:"long_url,omitempty"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func (r UpdateURLRequest) HasUpdates() bool {
	return r.LongURL != nil || r.CustomAlias != nil || r.ExpiresAt != nil
}
