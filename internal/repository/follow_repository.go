package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"
)

type FollowStore interface {
	Create(ctx context.Context, follow *models.Follow) (*models.Follow, error)
	GetFollowings(ctx context.Context, followerID int64) ([]*models.Follow, error)
	GetFollowers(ctx context.Context, followingID int64) ([]*models.Follow, error)
	GetRelationship(ctx context.Context, userA, userB int64) (*models.RelationShip, error)
	Delete(ctx context.Context, followerID, followingID int64) error
}

type FollowCache interface {
	AddFollowing(ctx context.Context, followerID, followingID int64) error
	IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error)
	GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error)
    GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
	InvalidatePair(ctx context.Context, followerID, followingID int64) error
	IInvalidateSelf(ctx context.Context, userID int64) error
}
type FollowRepository struct {
	followStore FollowStore
	followCache FollowCache
	sf          singleflight.Group
}

func NewFollowRepository(fs FollowStore, fc FollowCache) *FollowRepository {
	return &FollowRepository{
		followStore: fs,
		followCache: fc,
	}
}

func (r *FollowRepository) Create(ctx context.Context, followerID, followingID int64) (*dto.FollowRecord, error) {
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

	_ = r.followCache.InvalidatePair(ctx, followerID,followingID)

	return dto.NewFollowRecord(dbFollow), nil
}	

func(r *FollowRepository) CheckRelation(ctx context.Context, followerID, followingID int64) (*dto.RelationRecord, error) {
	if followerID == followingID {
		return &dto.RelationRecord{
            Following:  false,
            FollowedBy: false,
            IsMutual:   false,
        }, nil
	}

	low, high := followerID, followingID
	if low > high {
		low, high = high, low
	}

	sfKey := fmt.Sprintf("rel:%d:%d", low, high)

	ch := r.sf.DoChan(sfKey, func() (interface{}, error){
		return r.followStore.GetRelationship(context.Background(), followerID, followingID)
	})
	
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res, ok := <-ch:
        if !ok {
			return nil, errcode.ErrInternal
		}

		if res.Err != nil {
			return nil, res.Err
		}

		return dto.NewRelationRecord(res.Val.(*models.RelationShip)), nil
	}
}