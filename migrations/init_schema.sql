BEGIN;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name VARCHAR(120) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CHECK (char_length(btrim(email)) > 0),
    CHECK (email = lower(email)),
    CHECK (char_length(btrim(name)) > 0)
);

CREATE TABLE urls (
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

-- expired links for cleanup jobs without indexing non-expiring URLs.
CREATE INDEX urls_expires_at_idx ON urls (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX urls_user_created_at_idx ON urls (user_id, created_at DESC);

CREATE TABLE refresh_tokens (
    -- Stores a SHA-256 hash of the raw token returned to the client.
    token CHAR(64) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,

    CHECK (expires_at > created_at)
);

CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_active_expiry_idx ON refresh_tokens (expires_at) WHERE revoked_at IS NULL;

COMMIT;
