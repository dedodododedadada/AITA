ALTER TABLE users 
ADD COLUMN follower_count BIGINT NOT NULL DEFAULT 0,
ADD COLUMN following_count BIGINT NOT NULL DEFAULT 0;

ALTER TABLE users
ADD CONSTRAINT follower_count_min CHECK (follower_count >= 0),
ADD CONSTRAINT following_count_min CHECK (following_count >= 0);
