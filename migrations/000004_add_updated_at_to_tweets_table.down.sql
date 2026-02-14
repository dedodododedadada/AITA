DROP TRIGGER IF EXISTS update_tweet_modtime ON tweets;

DROP FUNCTION IF EXISTS update_modified_column();

ALTER TABLE tweets DROP COLUMN IF EXISTS updated_at;