CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    short_code VARCHAR(64) NOT NULL UNIQUE,
    long_url TEXT NOT NULL,
    custom_alias VARCHAR(64) UNIQUE,
    click_count BIGINT NOT NULL DEFAULT 0 CHECK (click_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,

    CHECK (char_length(btrim(short_code)) > 0),
    CHECK (char_length(btrim(long_url)) > 0),
    CHECK (expires_at IS NULL OR expires_at > created_at)
);

CREATE INDEX IF NOT EXISTS urls_expires_at_idx ON urls (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS urls_user_created_at_idx ON urls (user_id, created_at DESC);
