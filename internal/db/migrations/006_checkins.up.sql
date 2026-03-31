DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'checkin_status') THEN
        CREATE TYPE checkin_status AS ENUM ('done', 'missed', 'late');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS check_ins (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID           NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date         DATE           NOT NULL,
    status       checkin_status NOT NULL,
    comment      TEXT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (challenge_id, user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_check_ins_challenge_user ON check_ins(challenge_id, user_id);