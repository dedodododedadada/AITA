package cache

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisTweetCache struct {
	client *redis.Client
	prefix string
}

func NewRedisTweetCache(c *redis.Client) *redisTweetCache {
	return &redisTweetCache{
		client: c,
		prefix: "tweet:",
	}
}

func (c *redisTweetCache) tweetKey(tweetID int64) string {
	return fmt.Sprintf("%scontent:%d", c.prefix, tweetID)
}

func(c *redisTweetCache) SetTweet(ctx context.Context, tweet *models.Tweet) error {
	keyTweet := c.tweetKey(tweet.ID)

	data, err := json.Marshal(tweet)
	if err != nil {
		slog.Error("[Redis Error] ツイートのシリアライズに失敗しました",
			"tweet_id", tweet.ID,
			"err", err,
		)
		return err
	}

	ttl := utils.GetRandomExpiration(12*time.Hour, 6*time.Hour)
	err = c.client.Set(ctx, keyTweet, data, ttl).Err()
	if err != nil {
		slog.Error("[Redis Error] ツイートキャッシュの保存に失敗しました",
			"tweet_id", tweet.ID,
			"err", err,
		)
		return err
	}	

	return nil
}

func(c *redisTweetCache) GetTweet(ctx context.Context, tweetID int64) (*models.Tweet, error) {
	keyTweet := c.tweetKey(tweetID)

	val, err := c.client.Get(ctx, keyTweet).Bytes()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, redis.Nil
		}
		slog.Warn("[Redis Error] ツイートキャッシュの取得に失敗しました",
			"tweet_id", tweetID,
			"err", err,
		)
		return nil, err
	}

	var tweet models.Tweet
	if err := json.Unmarshal(val, &tweet); err != nil {
		slog.Warn("[Redis Decode Error] ツイートのデコードに失敗しました",
			"tweet_id", tweetID,
			"err", err,
		)
		return nil, err
	}

	return &tweet, nil
}

func (c *redisTweetCache) MultiGetTweets(ctx context.Context, tweetIDs []int64) (map[int64]*models.Tweet, error) {
 	results := make(map[int64]*models.Tweet, len(tweetIDs))

	if len(tweetIDs) == 0 {
		return results, nil
	}

	pipe := c.client.Pipeline()

	cmds := make(map[int64]*redis.StringCmd, len(tweetIDs)) 
	for _, id := range tweetIDs {
		if id <= 0 {
			continue
		}

		cmds[id] = pipe.Get(ctx, c.tweetKey(id))
	}

	_, err := pipe.Exec(ctx)

	if err != nil  {
		 if errors.Is(err, redis.Nil) {
            return nil, redis.Nil
        }
		slog.Error("[Redis Error] ツイートリストの一括取得に失敗しました",
			"count", len(tweetIDs),
			"err", err,
		)
		return nil, err
	}

	for id, cmd := range cmds {
		dataBytes, err := cmd.Bytes()
		if err != nil {
			continue
		}

		var tweet models.Tweet
		if err := json.Unmarshal(dataBytes, &tweet); err != nil {
			slog.Warn("[Redis Decode Error] ツイートのデコードに失敗しました",
				"tweet_id", id,
				"err", err,
			)
			continue
		}

		results[id] = &tweet
	}

	return results, nil
}

func (c *redisTweetCache) Invalidate(ctx context.Context, tweetID int64) error {
	keyTweet := c.tweetKey(tweetID)

	err := c.client.Del(ctx, keyTweet).Err()

	if err != nil {
		slog.Error("[Redis Error] ツイートキャッシュの無効化に失敗しました",
            "tweet_id", tweetID,
            "key", keyTweet,
            "err", err,
        )
	}
	
	return err
}

func (c *redisTweetCache) MultiSetTweets(ctx context.Context, tweets []*models.Tweet) error {
    if len(tweets) == 0 {
        return nil
    }

    pipe := c.client.Pipeline()

	marshalCount := 0

    for _, tweet := range tweets {
        key := c.tweetKey(tweet.ID)
        data, err := json.Marshal(tweet)
		if err != nil {
			slog.Warn("[Redis Error] ツイートのシリアライズに失敗しました（一括処理中）",
				"tweet_id", tweet.ID,
				"err", err,
			)
			continue
		}
        ttl := utils.GetRandomExpiration(24*time.Hour, 3*time.Hour)
       	
		pipe.SetNX(ctx, key, data, ttl)
		marshalCount++
    }

	if marshalCount == 0 {
		return nil
	}

    _, err := pipe.Exec(ctx)
    if err != nil {
		slog.Error("[Redis Error] ツイートリストの一括保存に失敗しました",
			"requested_count", len(tweets),
			"marshal_count", marshalCount,
			"err", err,
		)
		return err
	}

	return nil
}
