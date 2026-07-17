-- +goose Up
-- One refresh token per device/session instead of one per user,
-- so logging in on a second device no longer logs out the first.
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- The single-token columns are superseded by the table above.
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token_hash;
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token_expires_at;

-- +goose Down
ALTER TABLE users ADD COLUMN refresh_token_hash TEXT;
ALTER TABLE users ADD COLUMN refresh_token_expires_at TIMESTAMPTZ;
DROP TABLE IF EXISTS refresh_tokens;
