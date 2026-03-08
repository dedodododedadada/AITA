package cache

import (
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisTimelineCache struct {
	client *redis.Client
	prefix string
}

func NewRedisTimelineCache(c *redis.Client) *redisTimelineCache {
	return &redisTimelineCache{
		client: c,
		prefix: "timeline:",
	}
}

func (c *redisTimelineCache) timelineKey(userID int64) string {
	return fmt.Sprintf("%s%d", c.prefix, userID)
}

func (c *redisTimelineCache) PushBatch(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error {
	pipe := c.client.Pipeline()
	score := float64(createdAt.Unix())

	for _, id := range userIDs {
		tlKey := c.timelineKey(id)
		pipe.ZAdd(ctx, tlKey, redis.Z{Score: score, Member: tweetID})
		pipe.ZRemRangeByRank(ctx, tlKey, 0, -1001)
		ttl := utils.GetRandomExpiration(72, 3)
		pipe.Expire(ctx, tlKey, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		slog.Error("[Redis Error] タイムラインの一括プッシュに失敗しました",
			"tweet_id", tweetID,
			"user_count", len(userIDs),
			"err", err,
		)
		return err
	}

	return nil
}

func (c *redisTimelineCache) FindRange(ctx context.Context, userID int64, start, stop int64) ([]int64, error) {
	if  start >= stop {
		return []int64{}, nil
	}

	tlKey := c.timelineKey(userID)
	res, err := c.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key: tlKey,
		Start: start,
		Stop: stop,
		Rev: true,
	}).Result()
	if err != nil {
		slog.Error("[Redis Error] タイムラインの取得に失敗しました",
			"user_id", userID,
			"key", tlKey,
			"start", start,
			"stop", stop,
			"err", err,
		)
		return nil, err
	}

	tweetIDs := make([]int64, 0, len(res)) 
	for _, idStr := range res {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		tweetIDs = append(tweetIDs, id)
	}

	return tweetIDs, nil
}


func(c *redisTimelineCache) RecallTweet(ctx context.Context, tweetID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	load := min(len(userIDs), 1000)
	batch := make([]int64, load)
	copy(batch, userIDs[:load])

	member := strconv.FormatInt(tweetID, 10)
	
	
	pipe := c.client.Pipeline()

	for _, id := range batch {
		tlKey := c.timelineKey(id)
		pipe.ZRem(ctx, tlKey, member)
	}

	_, err := pipe.Exec(ctx)

	if err != nil {
		slog.Error("[Redis Error] タイムラインからのツイート削除に失敗しました",
			"tweet_id", tweetID,
			"batch_size", len(batch),
			"err", err,
		)
		return err
	}

	return nil
}