ALTER TABLE tweets 

ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tweet_modtime
    BEFORE UPDATE ON tweets
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();