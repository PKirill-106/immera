-- +goose Up

CREATE TABLE auth_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    token_hash VARCHAR(64) NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_refresh_tokens_user_id
    ON auth_refresh_tokens(user_id);

CREATE INDEX idx_auth_refresh_tokens_expires_at
    ON auth_refresh_tokens(expires_at);

-- +goose Down

DROP TABLE IF EXISTS auth_refresh_tokens;