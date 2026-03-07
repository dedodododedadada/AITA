package cache

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	return fmt.Sprintf("%sdata:%d", c.prefix, userID)
}

func(c *redisUserCache) countKey(userID int64) string {
	return fmt.Sprintf("%scount:%d", c.prefix, userID)
}

func (c *redisUserCache) Add(ctx context.Context, info *models.UserInfo, fllwrCount, fllwngCount int64) error {
	dkey := c.dataKey(info.ID)
	cKey := c.countKey(info.ID)
	data, _ := json.Marshal(info)

	ttl := utils.GetRandomExpiration(72*time.Hour, 3*time.Hour)
	
	pipe := c.client.Pipeline()
	pipe.Set(ctx, dkey, data, ttl)
	pipe.HSet(ctx, cKey, map[string]interface{}{
		"follower": fllwrCount,
		"following": fllwngCount,
	})
	
	pipe.Expire(ctx, cKey, ttl - 30*time.Minute)

	_, err := pipe.Exec(ctx)

	if err != nil {
        slog.Error("[Redis Error] ユーザーIDのキャッシュ追加に失敗しました", 
			"user_id", info.ID, 
			"err", err,
		)
    }

	return err
}

func (c *redisUserCache) AddInfoOnly(ctx context.Context, info *models.UserInfo) {
    dkey := c.dataKey(info.ID)
    data, err := json.Marshal(info)
	if err != nil {
        slog.Error("[Redis Error] ユーザー情報のシリアライズに失敗しました", "user_id", info.ID, "err", err)
        return
    }
    ttl := utils.GetRandomExpiration(72*time.Hour, 3*time.Hour)
    
	err = c.client.Set(ctx, dkey, data, ttl).Err()
    if err != nil {
        slog.Error("[Redis Error] ユーザー情報のみのキャッシュ追加に失敗しました", 
            "user_id", info.ID, 
            "err", err,
        )
    }
}

func (c *redisUserCache) Invalidate(ctx context.Context, userID int64)  {
	dkey := c.dataKey(userID)
	cKey := c.countKey(userID)

	pipe := c.client.Pipeline()

	pipe.Del(ctx, dkey)
	pipe.Del(ctx, cKey)

	_, err := pipe.Exec(ctx)
	
	if err != nil {
		slog.Error("[Redis Error] ユーザーIDのキャッシュ削除に失敗しました",
			"user_id", userID,
			"err", err,
		)
    }
}

func (c *redisUserCache) Get(ctx context.Context, userID int64) (*models.UserInfo, int64, int64, error) {
    dKey := c.dataKey(userID)
    cKey := c.countKey(userID)

    pipe := c.client.Pipeline()

    dataCmd := pipe.Get(ctx, dKey)
    countCmd := pipe.HGetAll(ctx, cKey)

    _, err := pipe.Exec(ctx)

	if err != nil {
        if errors.Is(err, redis.Nil) {
            return nil, 0, 0, redis.Nil
        }
        slog.Error("[Redis Error] ユーザーデータの取得に失敗しました",
            "user_id", userID,
            "err", err,
        )
        return nil, 0, 0, err
    }
    var info models.UserInfo
    dataBytes, err := dataCmd.Bytes()
    if err == nil {
   		if err := json.Unmarshal(dataBytes, &info); err != nil {
            slog.Warn("[Redis Decode Error] JSONのデコードに失敗しました",
                "user_id", userID,
                "err", err,
            )
        }
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
    err :=  c.client.HIncrBy(ctx, cKey, "follower", delta).Err()
	if err != nil {
        slog.Error("[Redis Error] フォロワー数のインクリメントに失敗しました",
            "user_id", userID,
            "delta", delta,
            "err", err,
        )
    }

	return err
}

func (c *redisUserCache) IncrFollowing(ctx context.Context, userID int64, delta int64) error {
    cKey := c.countKey(userID)
    err :=  c.client.HIncrBy(ctx, cKey, "following", delta).Err()
	if err != nil {
        slog.Error("[Redis Error] フォロウィング数のインクリメントに失敗しました",
            "user_id", userID,
            "delta", delta,
            "err", err,
        )
    }

	return err
}

func (c *redisUserCache) Exists(ctx context.Context, userID int64) (bool, error) {
    dKey := c.dataKey(userID)
    cKey := c.countKey(userID)

    n, err := c.client.Exists(ctx, dKey, cKey).Result()
    if err != nil {
        slog.Error("[Redis Error] ユーザーの存在確認に失敗しました", 
            "user_id", userID, 
            "err", err,
        )
        return false, err
    }

    return n == 2, nil
}

func(c *redisUserCache) GetLists(ctx context.Context, userIDs []int64) (map[int64]*models.UserInfo, error) {
	results := make(map[int64]*models.UserInfo, len(userIDs))

	if len(userIDs) == 0 {
		return results, nil
	}

	pipe := c.client.Pipeline()

	cmds := make(map[int64]*redis.StringCmd, len(userIDs))
	for _, id := range userIDs {
		if id <= 0 {
			continue
		}

		cmds[id] = pipe.Get(ctx, c.dataKey(id))
	}

	_, err := pipe.Exec(ctx)

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return results, redis.Nil
		}
        slog.Warn("[Redis Error] ユーザーリストの取得に失敗しました",
            "count", len(userIDs),
            "err", err,
        )
        return nil, err
    }

	for id, cmd := range cmds {
		dataBytes, err := cmd.Bytes()
		if err != nil {
			continue
		}

		var info models.UserInfo
		if err := json.Unmarshal(dataBytes, &info); err != nil {
			slog.Error("[Redis Decode Error] ユーザー情報のデコードに失敗しました",
				"user_id", id,
				"err", err,
			)
			continue
		}

		results[id] = &info
	}

	return results, nil
}


func (c *redisUserCache) AddLists(ctx context.Context, infos []*models.UserInfo) error {
    if len(infos) == 0 {
        return nil
    }

    pipe := c.client.Pipeline()
    count := 0

    for _, info := range infos {
        if info == nil { continue }

        data, err := json.Marshal(info)
        if err != nil {
            slog.Error("[Redis Error] ユーザー情報のシリアライズに失敗しました",
                "user_id", info.ID, "err", err,
            )
            continue
        }

        key := c.dataKey(info.ID)
        ttl := utils.GetRandomExpiration(72*time.Hour, 3*time.Hour)

        pipe.SetNX(ctx, key, data, ttl)
        count++
    }

    if count == 0 { return nil }

    _, err := pipe.Exec(ctx)
    if err != nil {
        slog.Error("[Redis Error] ユーザーリストのキャッシュ一括追加に失敗しました",
            "count", len(infos), "err", err,
        )
        return err
    }
    return nil
}