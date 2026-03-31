CREATE TABLE IF NOT EXISTS categories (
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT NULL UNIQUE
);

INSERT INTO categories (name) VALUES
    ('Sport'),
    ('Study'),
    ('Health'),
    ('Finance'),
    ('Other')
ON CONFLICT (name) DO NOTHING;