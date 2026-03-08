package cache

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var addFollowLua = redis.NewScript(`
    local added = 0
    if redis.call("EXISTS", KEYS[1]) == 1 then
        redis.call("ZADD", KEYS[1], ARGV[1], ARGV[2])
        redis.call("EXPIRE", KEYS[1], ARGV[4])
        added = added + 1
    end
    if redis.call("EXISTS", KEYS[2]) == 1 then
        redis.call("ZADD", KEYS[2], ARGV[1], ARGV[3])
        redis.call("EXPIRE", KEYS[2], ARGV[4])
        added = added + 1
    end
    return added`)


type redisFollowCache struct {
	client *redis.Client
	prefix string
}

func NewRedisFollowCache(c *redis.Client) *redisFollowCache {
	return &redisFollowCache{
		client: c,
		prefix: "follow:",
	}
}

func (c *redisFollowCache) followingKey(userID int64) string {
	return fmt.Sprintf("%sfollowing:%d", c.prefix, userID)
}
func (c *redisFollowCache) followerKey(userID int64) string {
	return fmt.Sprintf("%sfollower:%d", c.prefix, userID)
}

func (c *redisFollowCache) Add(ctx context.Context, followerID, followingID int64, score float64) error {
    keyFollowing := c.followingKey(followerID)
    keyFollower := c.followerKey(followingID)
    ttl := int(utils.GetRandomExpiration(24*time.Hour, 1*time.Hour).Seconds())

    _, err := addFollowLua.Run(ctx, c.client, 
        []string{keyFollowing, keyFollower}, 
        score, followingID, followerID, ttl,
    ).Result()

    if err != nil {
        slog.Error("[Redis Lua Error] 原子フォロー追加失敗", "err", err)
    }
    return err
}

func(c *redisFollowCache) AddFollowings(ctx context.Context, followerID int64, sets []*models.CacheMember) error {
	keyFollowing := c.followingKey(followerID)
	if len(sets) == 0 {
		return nil
	}

	zMembers := make([]redis.Z, len(sets))
    for i, m := range sets {
        zMembers[i] = redis.Z{
            Score:  m.Score,
            Member: m.Member,
        }
    }

	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
	
	pipe := c.client.Pipeline()
	pipe.ZAdd(ctx, keyFollowing, zMembers...)
	pipe.Expire(ctx,keyFollowing, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		slog.Error("[Redis Error] フォロー中リストのキャッシュ一括追加に失敗しました",
			"user_id", followerID,
			"count", len(sets),
			"err", err,
		)
	}
	return err
}

func(c *redisFollowCache) AddFollowers(ctx context.Context, followingID int64, sets []*models.CacheMember) error {
	keyFollower := c.followerKey(followingID)

	pipe := c.client.Pipeline()

	zMembers := make([]redis.Z, len(sets))
    for i, m := range sets {
        zMembers[i] = redis.Z{
            Score:  m.Score,
            Member: m.Member,
        }
    }

	pipe.ZAdd(ctx, keyFollower, zMembers...)
	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
	pipe.Expire(ctx,keyFollower, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		slog.Error("[Redis Error] フォロワーリストのキャッシュ一括追加に失敗しました",
			"user_id", followingID,
			"count", len(sets),
			"err", err,
		)
	}
	return err
}

// When checkFollowing == true ,user as follower  When checkFollowing == false, user as following
func(c *redisFollowCache) Exists(ctx context.Context, userID int64, checkFollowing bool) (bool, error) {
	key := c.followingKey(userID)
	
	if !checkFollowing {
		key = c.followerKey(userID)
	} 
	
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		slog.Warn("[Redis Error] キャッシュの存在確認に失敗しました",
			"user_id", userID,
			"check_following", checkFollowing,
			"err", err,
		)
		return false, err
	}

	return n > 0, nil
}

func (c *redisFollowCache) GetRelation(ctx context.Context, userID, targetID int64) (isFollowing, isFollowed bool, err error) {
    keyFollowing := c.followingKey(userID)
	keyFollower := c.followerKey(userID)
	
	pipe := c.client.Pipeline()
    
	exFollowingCmd := pipe.Exists(ctx, keyFollowing)
    exFollowerCmd := pipe.Exists(ctx, keyFollower)

	targetStr := strconv.FormatInt(targetID, 10)
    fCmd := pipe.ZScore(ctx, keyFollowing, targetStr)
    tCmd := pipe.ZScore(ctx, keyFollower, targetStr)
    
    _, _ = pipe.Exec(ctx)

	if exFollowingCmd.Val() == 0 || exFollowerCmd.Val() == 0 {
        return false, false, redis.Nil
    }
	
    isFollowing = fCmd.Err() == nil
    isFollowed = tCmd.Err() == nil

    return isFollowing, isFollowed, nil
}

func (c *redisFollowCache) FindFollowingIDs(ctx context.Context, userID int64) ([]int64, error) {
	key := c.followingKey(userID)
	return c.findIDsFromZSet(ctx, key)
}
func (c *redisFollowCache) FindFollowerIDs(ctx context.Context, userID int64) ([]int64,  error) {
	key := c.followerKey(userID)
	return c.findIDsFromZSet(ctx, key)
}
func (c *redisFollowCache) findIDsFromZSet(ctx context.Context, key string) ([]int64,  error) {
	strs, err := c.client.ZRangeArgs(ctx, redis.ZRangeArgs{
        Key:   key,
        Start: 0,
        Stop:  -1,
        Rev:   true, 
    }).Result()


	if err != nil {
		slog.Error("[Redis Error] IDリストの取得に失敗しました", "key", key, "err", err)
		return nil, err
	}

	ids := make([]int64, 0, len(strs))
	for  _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64) 
		if err != nil {
			slog.Warn("[Redis Data Error] IDのパースに失敗しました", "value", s, "err", err)
			continue
		}

		ids = append(ids, id)
	}
	return ids, nil
}

func (c *redisFollowCache) InvalidatePair(ctx context.Context, followerID, followingID int64) error {
    keyFollowing := c.followingKey(followerID)
	keyFollower:= c.followerKey(followingID)

	pipe := c.client.Pipeline()

	pipe.Del(ctx, keyFollowing)
    pipe.Del(ctx, keyFollower)
   
	_, err := pipe.Exec(ctx)
	if err != nil {
		slog.Error("[Redis Error] 関係キャッシュの削除に失敗しました",
			"follower_id", followerID,
			"following_id", followingID,
			"err", err,
		)
	}
	return err
}

func (c *redisFollowCache) InvalidateSelf(ctx context.Context, userID int64,  isFollowing, isFollower bool) error {
   
	pipe := c.client.Pipeline()

  	if isFollowing {
        pipe.Del(ctx, c.followingKey(userID))
    }
    if isFollower {
        pipe.Del(ctx, c.followerKey(userID))
    }

    _, err := pipe.Exec(ctx)
	if err != nil {
		slog.Error("[Redis Error] 自己キャッシュの削除に失敗しました",
			"user_id", userID,
			"err", err,
		)
	}
    return err
}