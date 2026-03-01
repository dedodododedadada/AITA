package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"
)

// follow unfollow relation（without rel key in redis）完成
type FollowStore interface {
	Create(ctx context.Context, follow *models.Follow) (*models.Follow, error)
	GetFollowings(ctx context.Context, followerID int64) ([]*models.Follow, error)
	GetFollowers(ctx context.Context, followingID int64) ([]*models.Follow, error)
	GetRelationship(ctx context.Context, userA, userB int64) (*models.RelationShip, error)
	Delete(ctx context.Context, followerID, followingID int64) error
}

type FollowCache interface {
	Add(ctx context.Context, followerID, followingID int64) error 
	AddFollowings(ctx context.Context, followerID int64, followings []int64) error
	AddFollowers(ctx context.Context, followingID int64, followers []int64) error
	Exists(ctx context.Context, userID int64, IsFollowing bool) (bool, error)
	GetRelation(ctx context.Context, followerID, followingID int64) (isFollowing, isFollowed bool, err error)
	GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error)
    GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
	InvalidatePair(ctx context.Context, followerID, followingID int64) error
	InvalidateSelf(ctx context.Context, userID int64) error
}
type followRepository struct {
	followStore FollowStore
	followCache FollowCache
	sfFollow    *singleflight.Group
}

func NewFollowRepository(fs FollowStore, fc FollowCache) *followRepository {
	return &followRepository{
		followStore: fs,
		followCache: fc,
	}
}

func (r *followRepository) Create(ctx context.Context, followerID, followingID int64) (*dto.FollowRecord, error) {
	if followerID == followingID {
		return nil, errcode.ErrCannotFollowSelf
	}

	dbFollow, err := r.followStore.Create(ctx, &models.Follow{
		FollowerID: followerID,
		FollowingID: followingID,
	})

	if err != nil {
		return nil, err
	}

	okFollowings, _ := r.followCache.Exists(ctx, followerID, true) 
	okFollowers, _ := r.followCache.Exists(ctx, followingID, false)

	if okFollowers && okFollowings  {
		_ = r.followCache.Add(ctx, followerID, followingID)
	} else {
		_ = r.followCache.InvalidatePair(ctx, followerID,followingID)
	}
	
	return dto.NewFollowRecord(dbFollow), nil
}	

func(r *followRepository) CheckRelation(ctx context.Context, followerID, followingID int64) (*dto.RelationRecord, error) {
	if followerID == followingID {
		return &dto.RelationRecord{}, nil
	}

	okFollowing, _ := r.followCache.Exists(ctx, followerID, true)
    okFollower, _:= r.followCache.Exists(ctx, followingID, false)

	if okFollowing && okFollower {
		isFollowing, isFollowed, err := r.followCache.GetRelation(ctx, followerID, followingID)

		if err == nil {
			return &dto.RelationRecord{
				Following: isFollowing,
				FollowedBy: isFollowed,
				IsMutual: isFollowing && isFollowed,
			}, err
		}
	}
	
	low, high := followerID, followingID
	if low > high {
		low, high = high, low
	}

	sfKey := fmt.Sprintf("rel:%d:%d", low, high)

	res, err := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func() (*models.RelationShip, error) {
		return r.followStore.GetRelationship(context.Background(), followerID, followingID)
	})

	if err != nil {
		return nil, err
	}

	return dto.NewRelationRecord(res), nil
}

func(r *followRepository) GetFollowings(ctx context.Context, userID int64) ([]int64, error) {
	list, err := r.followCache.GetFollowingIDs(ctx, userID)
    if err == nil && len(list) > 0 {
        return list, nil
    }

	sfKey := fmt.Sprintf("followings:%d", userID)

	followings, dberr := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func() ([]*models.Follow, error) {
		return r.followStore.GetFollowings(context.Background(), userID)
	})

	if dberr != nil {
		return nil, dberr
	}
	ids := make([]int64, len(followings))

	for i, f := range followings {
		ids[i] = f.FollowerID
	}

	go func() {
		_ = r.followCache.AddFollowings(context.Background(), userID, ids)
	}()
	
	return ids, nil
}

func(r *followRepository) GetFollowers(ctx context.Context, userID int64) ([]int64, error) {
	list, err := r.followCache.GetFollowerIDs(ctx, userID)
    if err == nil && len(list) > 0 {
        return list, nil
    }

	sfKey := fmt.Sprintf("followers:%d", userID)

	followers, dberr := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func() ([]*models.Follow, error) {
		return r.followStore.GetFollowers(context.Background(), userID)
	})

	if dberr != nil {
		return nil, dberr
	}
	ids := make([]int64, len(followers))

	for i, f := range followers {
		ids[i] = f.FollowingID
	}

	go func() {
		_ = r.followCache.AddFollowers(context.Background(), userID, ids)
	}()
	
	return ids, nil
}



func (r *followRepository) RemoveFollow(ctx context.Context, followerID, followingID int64) error {
	if (followerID == followingID) {
		return nil
	}

	err := r.followStore.Delete(ctx, followerID, followingID)
	if err != nil {
		return err
	}

	_ = r.followCache.InvalidatePair(ctx, followerID, followingID)
	return nil
}