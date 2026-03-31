DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'challenge_status') THEN
        CREATE TYPE challenge_status AS ENUM ('upcoming', 'active', 'finished');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS challenges (
    id            UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id    UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id   INT              NOT NULL REFERENCES categories(id),
    title         TEXT             NOT NULL,
    description   TEXT,
    starts_at     DATE             NOT NULL,
    ends_at       DATE             NOT NULL,
    working_days  INT[]            NOT NULL DEFAULT '{0,1,2,3,4,5,6}',
    max_skips     INT              NOT NULL DEFAULT 0,
    deadline_time TIME             NOT NULL DEFAULT '23:00',
    is_public     BOOLEAN          NOT NULL DEFAULT false,
    invite_token  UUID             UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    status        challenge_status NOT NULL DEFAULT 'upcoming',
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_challenges_creator_id ON challenges(creator_id);
CREATE INDEX IF NOT EXISTS idx_challenges_status     ON challenges(status);
CREATE INDEX IF NOT EXISTS idx_challenges_is_public  ON challenges(is_public);