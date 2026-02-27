ALTER TABLE users 
DROP COLUMN IF EXISTS follower_count,
DROP COLUMN IF EXISTS following_count;