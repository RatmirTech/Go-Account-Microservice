CREATE TABLE IF NOT EXISTS refresh_tokens (
                                              id UUID PRIMARY KEY,
                                              user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    device_id TEXT NOT NULL,
    ip TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
    );

CREATE INDEX idx_refresh_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_device_id ON refresh_tokens(device_id);