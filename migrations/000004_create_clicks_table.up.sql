CREATE TABLE IF NOT EXISTS clicks (
    id UUID PRIMARY KEY,
    url_id UUID NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    short_code VARCHAR(64) NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    country VARCHAR(64),
    city VARCHAR(128),
    device_type VARCHAR(32) NOT NULL DEFAULT 'unknown',
    browser VARCHAR(32) NOT NULL DEFAULT 'Other',
    os VARCHAR(32) NOT NULL DEFAULT 'Other',
    referrer VARCHAR(512),
    referrer_host VARCHAR(128),
    ip_hash CHAR(64) NOT NULL,

    CHECK (char_length(btrim(short_code)) > 0),
    CHECK (char_length(btrim(ip_hash)) > 0)
);

CREATE INDEX IF NOT EXISTS clicks_url_id_clicked_at_idx ON clicks (url_id, clicked_at DESC);
CREATE INDEX IF NOT EXISTS clicks_short_code_clicked_at_idx ON clicks (short_code, clicked_at DESC);
CREATE INDEX IF NOT EXISTS clicks_url_id_country_idx ON clicks (url_id, country);
CREATE INDEX IF NOT EXISTS clicks_url_id_device_idx ON clicks (url_id, device_type);
CREATE INDEX IF NOT EXISTS clicks_url_id_browser_idx ON clicks (url_id, browser);
CREATE INDEX IF NOT EXISTS clicks_url_id_os_idx ON clicks (url_id, os);
CREATE INDEX IF NOT EXISTS clicks_url_id_referrer_host_idx ON clicks (url_id, referrer_host);
