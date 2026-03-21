CREATE TABLE IF NOT EXISTS checkins (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID        NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date         DATE        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (challenge_id, user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_checkins_challenge_user ON checkins(challenge_id, user_id);
CREATE INDEX IF NOT EXISTS idx_checkins_challenge_date ON checkins(challenge_id, date);