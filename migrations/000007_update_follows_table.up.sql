DROP INDEX IF EXISTS idx_follows_following_id;

DROP TABLE IF EXISTS follows;

CREATE TABLE follows (
    id           BIGSERIAL PRIMARY KEY, 
    follower_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE (follower_id, following_id)
);

CREATE INDEX IF NOT EXISTS idx_follows_following_id ON follows(following_id);