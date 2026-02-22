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

func (c *redisFollowCache) Add(ctx context.Context, followerID, followingID int64) error {
    keyFollowing := c.followingKey(followerID)
	keyFollower := c.followerKey(followingID)
	
	pipe := c.client.Pipeline()
    
	pipe.SAdd(ctx, keyFollowing, followingID)
    pipe.SAdd(ctx, keyFollower, followerID)
    
	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
    pipe.Expire(ctx, keyFollowing , expiration)
    pipe.Expire(ctx, keyFollower, expiration)

    _, err := pipe.Exec(ctx)
    return err
}

func(c *redisFollowCache) AddFollowings(ctx context.Context, followerID int64, followings []int64) error {
	keyFollowing := c.followingKey(followerID)

	pipe := c.client.Pipeline()

	pipe.SAdd(ctx, keyFollowing, followings)
	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
	pipe.Expire(ctx,keyFollowing, expiration)

	_, err := pipe.Exec(ctx)
	return err
}

func(c *redisFollowCache) AddFollowers(ctx context.Context, followingID int64, followers []int64) error {
	keyFollower := c.followingKey(followingID)

	pipe := c.client.Pipeline()

	pipe.SAdd(ctx, keyFollower, followers)
	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
	pipe.Expire(ctx,keyFollower, expiration)

	_, err := pipe.Exec(ctx)
	return err
}

// When IsFollowing == true ,user as follower  When IsFollowing == false, user as following
func(c *redisFollowCache) Exists(ctx context.Context, userID int64, IsFollowing bool) (bool, error) {
	key := c.followingKey(userID)
	
	if !IsFollowing {
		key = c.followerKey(userID)
	} 
	
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

func (c *redisFollowCache) GetRelation(ctx context.Context, followerID, followingID int64) (isFollowing, isFollowed bool, err error) {
    keyFollowing := c.followingKey(followerID)
	keyFollower := c.followerKey(followingID)
	
	pipe := c.client.Pipeline()
    
    fCmd := pipe.SIsMember(ctx, keyFollowing, followingID)
    tCmd := pipe.SIsMember(ctx, keyFollower, followerID)
    
    _, err = pipe.Exec(ctx)
    if err != nil {
        return false, false, err
    }
    
    return fCmd.Val(), tCmd.Val(), nil
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

func (c *redisFollowCache) InvalidatePair(ctx context.Context, followerID, followingID int64) error {
    keyFollowing := c.followingKey(followerID)
	keyFollower:= c.followerKey(followingID)

	pipe := c.client.Pipeline()

	pipe.Del(ctx, keyFollowing)
    pipe.Del(ctx, keyFollower)
   
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisFollowCache) InvalidateSelf(ctx context.Context, userID int64) error {
    keyFollowing := c.followingKey(userID)
    keyFollower := c.followerKey(userID)
   
	pipe := c.client.Pipeline()

    pipe.Del(ctx, keyFollowing)
    pipe.Del(ctx, keyFollower)

    _, err := pipe.Exec(ctx)
    return err
}