package cache

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

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
	
	pipe := c.client.Pipeline()
    
	pipe.ZAdd(ctx, keyFollowing, redis.Z{Score: score, Member: followingID})
    pipe.ZAdd(ctx, keyFollower, redis.Z{Score:score, Member: followerID})
    
	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
    pipe.Expire(ctx, keyFollowing , expiration)
    pipe.Expire(ctx, keyFollower, expiration)

    _, err := pipe.Exec(ctx)
    return err
}

func(c *redisFollowCache) AddFollowings(ctx context.Context, followerID int64, sets []*models.CacheMember) error {
	keyFollowing := c.followingKey(followerID)
	if len(sets) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()

	zMembers := make([]redis.Z, len(sets))
    for i, m := range sets {
        zMembers[i] = redis.Z{
            Score:  m.Score,
            Member: m.Member,
        }
    }

	pipe.ZAdd(ctx, keyFollowing, zMembers...)
	expiration := utils.GetRandomExpiration(24*time.Hour, 1*time.Hour)
	pipe.Expire(ctx,keyFollowing, expiration)

	_, err := pipe.Exec(ctx)
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
		return false, err
	}

	return n > 0, nil
}

func (c *redisFollowCache) GetRelation(ctx context.Context, followerID, followingID int64) (isFollowing, isFollowed bool, err error) {
    keyFollowing := c.followingKey(followerID)
	keyFollower := c.followerKey(followingID)
	
	pipe := c.client.Pipeline()
    
    fCmd := pipe.ZScore(ctx, keyFollowing, strconv.FormatInt(followingID, 10))
    tCmd := pipe.ZScore(ctx, keyFollower, strconv.FormatInt(followerID, 10))
    
    _, err = pipe.Exec(ctx)
    if err != nil {
        return false, false, err
    }
    
    return fCmd.Val() > 0, tCmd.Val() > 0, nil
}

func (c *redisFollowCache) GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error) {
	key := c.followingKey(userID)
	return c.getIDsFromZSet(ctx, key)
}
func (c *redisFollowCache) GetFollowerIDs(ctx context.Context, userID int64) ([]int64,  error) {
	key := c.followerKey(userID)
	return c.getIDsFromZSet(ctx, key)
}
func (c *redisFollowCache) getIDsFromZSet(ctx context.Context, key string) ([]int64,  error) {
	strs, err := c.client.ZRangeArgs(ctx, redis.ZRangeArgs{
        Key:   key,
        Start: 0,
        Stop:  -1,
        Rev:   true, 
    }).Result()
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

func (c *redisFollowCache) InvalidateSelf(ctx context.Context, userID int64,  isFollowing, isFollower bool) error {
   
	pipe := c.client.Pipeline()

  	if isFollowing {
        pipe.Del(ctx, c.followingKey(userID))
    }
    if isFollower {
        pipe.Del(ctx, c.followerKey(userID))
    }

    _, err := pipe.Exec(ctx)
    return err
}