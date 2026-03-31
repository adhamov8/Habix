-- Add comment field to checkins
ALTER TABLE checkins ADD COLUMN IF NOT EXISTS comment TEXT NOT NULL DEFAULT '';

-- Add data JSONB column to feed_events for storing extra event data
ALTER TABLE feed_events ADD COLUMN IF NOT EXISTS data JSONB;
