CREATE TABLE IF NOT EXISTS feed_comments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    feed_event_id UUID        NOT NULL REFERENCES feed_events(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text          TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_comments_event ON feed_comments(feed_event_id, created_at);
