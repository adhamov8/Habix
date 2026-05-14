DROP TABLE IF EXISTS feed_comments;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS user_badges;
DROP TABLE IF EXISTS badge_definitions;
DROP TABLE IF EXISTS checkins;
DROP TABLE IF EXISTS check_in_likes;
DROP TABLE IF EXISTS check_in_comments;
DROP TABLE IF EXISTS feed_events;
DROP TABLE IF EXISTS check_in_images;
DROP TABLE IF EXISTS check_ins;
DROP TABLE IF EXISTS challenge_participants;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS feed_event_type;
DROP TYPE IF EXISTS checkin_status;
DROP TYPE IF EXISTS challenge_status;

DROP EXTENSION IF EXISTS "pgcrypto";
