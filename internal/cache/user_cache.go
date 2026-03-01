package cache

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisUserCache struct {
	client *redis.Client
	prefix string

}

func NewRedisUserCache(c *redis.Client) *redisUserCache {
	return &redisUserCache{
		client: c,
		prefix: "user:",
	}
}

func (c *redisUserCache) dataKey(userID int64) string {
	return fmt.Sprintf("%s:data:%d", c.prefix, userID)
}

func(c *redisUserCache) countKey(userID int64)string {
	return fmt.Sprintf("%s:count:%d", c.prefix, userID)
}

func (c *redisUserCache) Add(ctx context.Context, info *models.UserCacheInfo, fllwrCount, fllwngCount int64) {
	dkey := c.dataKey(info.ID)
	cKey := c.countKey(info.ID)
	data, _ := json.Marshal(info)

	ttl := utils.GetRandomExpiration(72 * time.Hour, 1 * time.Hour)
	pipe := c.client.Pipeline()

	pipe.Set(ctx, dkey, data, ttl)
	pipe.HSet(ctx, cKey, map[string]interface{}{
		"follower": fllwrCount,
		"following": fllwrCount,
	})
	
	pipe.Expire(ctx, cKey, ttl - 30 * time.Minute)

	_, err := pipe.Exec(ctx)

	if err != nil {
        fmt.Printf("[Redis Error] ユーザーIDのキャッシュ追加に失敗しました %d: %v\n", info.ID, err)
    }
}

func(c *redisUserCache) Invalidate(ctx context.Context, userID int64)  {
	dkey := c.dataKey(userID)
	cKey := c.countKey(userID)

	pipe := c.client.Pipeline()

	pipe.Del(ctx, dkey)
	pipe.Del(ctx, cKey)

	_, err := pipe.Exec(ctx)
	
	if err != nil {
        fmt.Printf("[Redis Error] ユーザーIDのキャッシュ削除に失敗しました %d: %v\n", userID, err)
    }
}

func (c *redisUserCache) Get(ctx context.Context, userID int64) (*models.UserCacheInfo, int64, int64, error) {
    dKey := c.dataKey(userID)
    cKey := c.countKey(userID)

    pipe := c.client.Pipeline()

    dataCmd := pipe.Get(ctx, dKey)
    countCmd := pipe.HGetAll(ctx, cKey)

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, 0, 0, err
    }

    var info models.UserCacheInfo
    dataBytes, err := dataCmd.Bytes()
    if err == nil {
        _ = json.Unmarshal(dataBytes, &info)
    }
	
    counts := countCmd.Val()
    follower, _ := strconv.ParseInt(counts["follower"], 10, 64)
    following, _ := strconv.ParseInt(counts["following"], 10, 64)

    if dataBytes == nil || len(counts) < 2 {
        return nil, 0, 0, redis.Nil
    }

    return &info, follower, following, nil
}

func (c *redisUserCache) IncrFollower(ctx context.Context, userID int64, delta int64) error {
    cKey := c.countKey(userID)
    return c.client.HIncrBy(ctx, cKey, "follower", delta).Err()
}

func (c *redisUserCache) IncrFollowing(ctx context.Context, userID int64, delta int64) error {
    cKey := c.countKey(userID)
    return c.client.HIncrBy(ctx, cKey, "following", delta).Err()
}

func (c *redisUserCache) Exists(ctx context.Context, userID int64) (bool, error) {
    dKey := c.dataKey(userID)
    cKey := c.countKey(userID)

    n, err := c.client.Exists(ctx, dKey, cKey).Result()
    if err != nil {
        return false, err
    }

    return n == 2, nil
}