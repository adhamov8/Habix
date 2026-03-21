DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'feed_event_type') THEN
        CREATE TYPE feed_event_type AS ENUM ('challenge_created', 'user_joined', 'check_in');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS feed_events (
    id           UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID            NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         feed_event_type NOT NULL,
    reference_id UUID,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_events_challenge_id ON feed_events(challenge_id, created_at DESC);