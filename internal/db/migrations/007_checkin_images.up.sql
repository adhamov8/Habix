CREATE TABLE IF NOT EXISTS check_in_images (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    check_in_id UUID        NOT NULL REFERENCES check_ins(id) ON DELETE CASCADE,
    url         TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_check_in_images_check_in_id ON check_in_images(check_in_id);