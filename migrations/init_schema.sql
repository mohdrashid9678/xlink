BEGIN;

CREATE TABLE urls (
    id UUID PRIMARY KEY,
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

-- expired links for cleanup jobs without indexing non-expiring URLs.
CREATE INDEX urls_expires_at_idx ON urls (expires_at) WHERE expires_at IS NOT NULL;

COMMIT;
