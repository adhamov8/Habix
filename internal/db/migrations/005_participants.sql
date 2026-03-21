CREATE TABLE IF NOT EXISTS challenge_participants (
    challenge_id UUID        NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (challenge_id, user_id)
);