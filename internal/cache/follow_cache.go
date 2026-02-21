package cache

import (
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisFollowCache struct {
	client *redis.Client
}

func NewRedisFollowCache(c *redis.Client) *redisFollowCache {
	return &redisFollowCache{
		client: c,
	}
}

func (c *redisFollowCache) followingKey(userID int64) string {
	return fmt.Sprintf("follow:following:%d", userID)
}
func (c *redisFollowCache) followerKey(userID int64) string {
	return fmt.Sprintf("follow:follower:%d", userID)
}


func (c *redisFollowCache) AddFollowLink(ctx context.Context, followerID, followingID int64) error {
	keyFollowing := c.followingKey(followerID)
	keyFollower := c.followerKey(followingID)

	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)

	pipe := c.client.Pipeline()
	
	pipe.SAdd(ctx, keyFollowing, followingID)
	pipe.Expire(ctx, keyFollowing, expiration)
	
	pipe.SAdd(ctx, keyFollower, followerID)
	pipe.Expire(ctx, keyFollower, expiration)

	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisFollowCache) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
    key := c.followingKey(followerID)
    return c.client.SIsMember(ctx, key, followingID).Result()
}

func (c *redisFollowCache) GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error) {
	key := c.followingKey(userID)
	return c.getIDsFromSet(ctx, key)
}
func (c *redisFollowCache) GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error) {
	key := c.followerKey(userID)
	return c.getIDsFromSet(ctx, key)
}
func (c *redisFollowCache) getIDsFromSet(ctx context.Context, key string) ([]int64, error) {
	strs, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(strs) == 0 {
		return []int64{}, nil
	}

	ids := make([]int64, 0, len(strs))
	for  _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64) 
		if err != nil {
			continue
		}

		ids = append(ids, id)
	}
	return ids, nil
}

func (c *redisFollowCache) RemoveFollowing(ctx context.Context, followerID, followingID int64) error {
    keyFollowing := c.followingKey(followerID)
	keyFollower:= c.followerKey(followingID)

	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
	pipe := c.client.Pipeline()

	pipe.SRem(ctx, keyFollowing, followingID)
	pipe.Expire(ctx, keyFollowing, expiration)
	pipe.SRem(ctx, keyFollower, followerID)
	pipe.Expire(ctx, keyFollower, expiration)
   
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisFollowCache) Invalidate(ctx context.Context, userID int64) error {
    keyFollowing := c.followingKey(userID)
    keyFollower := c.followerKey(userID)
	expiration := utils.GetRandomExpiration(1*time.Hour, 10 *time.Minute)
    pipe := c.client.Pipeline()

    pipe.Expire(ctx, keyFollowing, expiration)
    pipe.Expire(ctx, keyFollower, expiration)

    _, err := pipe.Exec(ctx)
    return err
}