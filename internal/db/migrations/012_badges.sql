-- Badge definitions (reference table)
CREATE TABLE IF NOT EXISTS badge_definitions (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(50) UNIQUE NOT NULL,
    title       VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    icon        VARCHAR(10) NOT NULL
);

-- Seed badge definitions
INSERT INTO badge_definitions (code, title, description, icon) VALUES
    ('first_checkin',      'Первый шаг',        'Сделай свой первый check-in',                      '✅'),
    ('streak_3',           'Три дня подряд',     'Набери серию из 3 дней подряд',                    '🔥'),
    ('streak_7',           'Неделя огня',        'Набери серию из 7 дней подряд',                    '⚡'),
    ('streak_30',          'Месяц дисциплины',   'Набери серию из 30 дней подряд',                   '💎'),
    ('challenge_complete', 'Челлендж пройден',   'Заверши челлендж со 100% выполнением',             '🏆'),
    ('join_3_challenges',  'Активист',           'Участвуй в 3 и более челленджах',                  '🎯'),
    ('perfect_week',       'Идеальная неделя',   'Выполни все рабочие дни за одну календарную неделю','⭐')
ON CONFLICT (code) DO NOTHING;

-- User badges (awarded)
CREATE TABLE IF NOT EXISTS user_badges (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id     INT         NOT NULL REFERENCES badge_definitions(id) ON DELETE CASCADE,
    challenge_id UUID        REFERENCES challenges(id) ON DELETE SET NULL,
    earned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, badge_id, challenge_id)
);

CREATE INDEX IF NOT EXISTS idx_user_badges_user ON user_badges(user_id);
CREATE INDEX IF NOT EXISTS idx_user_badges_earned ON user_badges(earned_at DESC);

-- Add 'badge_earned' to feed_event_type enum
DO $$ BEGIN
    ALTER TYPE feed_event_type ADD VALUE IF NOT EXISTS 'badge_earned';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;