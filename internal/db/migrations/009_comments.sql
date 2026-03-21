CREATE TABLE IF NOT EXISTS check_in_comments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    check_in_id UUID        NOT NULL REFERENCES check_ins(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_check_in_comments_check_in_id ON check_in_comments(check_in_id);