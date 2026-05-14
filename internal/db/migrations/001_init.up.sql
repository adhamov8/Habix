CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'challenge_status') THEN
        CREATE TYPE challenge_status AS ENUM ('upcoming', 'active', 'finished');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'checkin_status') THEN
        CREATE TYPE checkin_status AS ENUM ('done', 'missed', 'late');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'feed_event_type') THEN
        CREATE TYPE feed_event_type AS ENUM ('challenge_created', 'user_joined', 'check_in', 'badge_earned');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    avatar_url    TEXT,
    bio           TEXT,
    timezone      TEXT        NOT NULL DEFAULT 'UTC',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id    ON refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS categories (
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT NULL UNIQUE
);

INSERT INTO categories (id, name) VALUES
    (1, 'Спорт и активность'),
    (2, 'Здоровье и питание'),
    (3, 'Учёба и саморазвитие'),
    (4, 'Профессиональные навыки'),
    (5, 'Творчество'),
    (6, 'Финансы'),
    (7, 'Ментальное здоровье'),
    (8, 'Отказ от вредных привычек'),
    (9, 'Другое')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;

-- Двигаем последовательность вперёд, чтобы новые id шли с 10
SELECT setval(pg_get_serial_sequence('categories', 'id'), 9, true);

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

CREATE TABLE IF NOT EXISTS challenge_participants (
    challenge_id UUID        NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (challenge_id, user_id)
);

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

CREATE TABLE IF NOT EXISTS check_in_images (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    check_in_id UUID        NOT NULL REFERENCES check_ins(id) ON DELETE CASCADE,
    url         TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_check_in_images_check_in_id ON check_in_images(check_in_id);

CREATE TABLE IF NOT EXISTS feed_events (
    id           UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID            NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         feed_event_type NOT NULL,
    reference_id UUID,
    data         JSONB,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_events_challenge_id ON feed_events(challenge_id, created_at DESC);

CREATE TABLE IF NOT EXISTS check_in_comments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    check_in_id UUID        NOT NULL REFERENCES check_ins(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_check_in_comments_check_in_id ON check_in_comments(check_in_id);

CREATE TABLE IF NOT EXISTS check_in_likes (
    check_in_id UUID        NOT NULL REFERENCES check_ins(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (check_in_id, user_id)
);

CREATE TABLE IF NOT EXISTS checkins (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID        NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date         DATE        NOT NULL,
    comment      TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (challenge_id, user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_checkins_challenge_user ON checkins(challenge_id, user_id);
CREATE INDEX IF NOT EXISTS idx_checkins_challenge_date ON checkins(challenge_id, date);

CREATE TABLE IF NOT EXISTS badge_definitions (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(50) UNIQUE NOT NULL,
    title       VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    icon        VARCHAR(10) NOT NULL
);

INSERT INTO badge_definitions (code, title, description, icon) VALUES
    ('first_checkin',      'Первая отметка',     'Сделай свою первую отметку',                        '✅'),
    ('streak_3',           'Серия 3 дня',         'Набери серию из 3 дней подряд',                     '🔥'),
    ('streak_7',           'Серия 7 дней',        'Набери серию из 7 дней подряд',                     '⚡'),
    ('streak_30',          'Серия 30 дней',       'Набери серию из 30 дней подряд',                    '💎'),
    ('challenge_complete', 'Идеальный результат', 'Завершил челлендж со 100% выполнением',             '🏆'),
    ('join_3_challenges',  'Активный участник',   'Участвует в 3 и более челленджах',                  '🎯'),
    ('perfect_week',       'Идеальная неделя',    'Выполни все рабочие дни за одну календарную неделю','⭐'),
    ('complete_50',        'Половина пути',       'Завершил челлендж с выполнением 50% и выше',        ''),
    ('complete_80',        'Хорошая работа',      'Завершил челлендж с выполнением 80% и выше',        ''),
    ('veteran',            'Ветеран',             'Завершил 5 и более челленджей',                     '')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_badges (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id     INT         NOT NULL REFERENCES badge_definitions(id) ON DELETE CASCADE,
    challenge_id UUID        REFERENCES challenges(id) ON DELETE SET NULL,
    earned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, badge_id, challenge_id)
);

CREATE INDEX IF NOT EXISTS idx_user_badges_user   ON user_badges(user_id);
CREATE INDEX IF NOT EXISTS idx_user_badges_earned ON user_badges(earned_at DESC);

CREATE TABLE IF NOT EXISTS notifications (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(50)  NOT NULL,
    title      VARCHAR(200) NOT NULL,
    body       TEXT         NOT NULL DEFAULT '',
    data       JSONB,
    is_read    BOOLEAN      NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user   ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;

CREATE TABLE IF NOT EXISTS feed_comments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    feed_event_id UUID        NOT NULL REFERENCES feed_events(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text          TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_comments_event ON feed_comments(feed_event_id, created_at);
