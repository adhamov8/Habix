CREATE TABLE IF NOT EXISTS check_in_likes (
    check_in_id UUID        NOT NULL REFERENCES check_ins(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (check_in_id, user_id)
);