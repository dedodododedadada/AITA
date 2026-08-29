CREATE INDEX IF NOT EXISTS idx_tweets_user_created_at ON tweets (user_id, created_at DESC);
