CREATE TABLE webauthn_credentials (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id),
    credential  BYTEA NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);