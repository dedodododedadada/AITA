package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
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
	Add(ctx context.Context, followerID, followingID int64, score float64) error
	AddFollowings(ctx context.Context, followerID int64, sets []*models.CacheMember) error
	AddFollowers(ctx context.Context, followingID int64, sets []*models.CacheMember) error
	Exists(ctx context.Context, userID int64, IsFollowing bool) (bool, error)
	GetRelation(ctx context.Context, followerID, followingID int64) (isFollowing, isFollowed bool, err error)
	GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error)
    GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
	InvalidatePair(ctx context.Context, followerID, followingID int64) error
	InvalidateSelf(ctx context.Context, userID int64,  isFollowing, isFollower bool) error
}
type followRepository struct {
	followStore FollowStore
	followCache FollowCache
	sfFollow    *singleflight.Group
	pool        *ants.Pool
}

func NewFollowRepository(fs FollowStore, fc FollowCache, p *ants.Pool) *followRepository {
	return &followRepository{
		followStore: fs,
		followCache: fc,
		sfFollow: &singleflight.Group{},
		pool: p,
	}
}

func makeScore(createdAt time.Time) float64 {
	return float64(createdAt.UTC().Unix())
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

	taskData := dbFollow
	err = r.pool.Submit(func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		okFollowings, _ := r.followCache.Exists(bgCtx, followerID, true) 
		okFollowers, _ := r.followCache.Exists(bgCtx, followingID, false)

		if okFollowers && okFollowings  {
			_ = r.followCache.Add(bgCtx, followerID, followingID, makeScore(taskData.CreatedAt))
		} else {
			_ = r.followCache.InvalidatePair(bgCtx, taskData.FollowerID, taskData.FollowingID)
		}
	})

	if err != nil {
   		slog.Warn("ants pool へのタスク投入に失敗しました。同期的なキャッシュ破棄を実行します。", "err", err)
		_ = r.followCache.InvalidatePair(context.Background(), followerID, followingID)
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

	res, err := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func(c context.Context) (*models.RelationShip, error) {
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

	followings, dberr := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func(c context.Context) ([]*models.Follow, error) {
		return r.followStore.GetFollowings(context.Background(), userID)
	})

	if dberr != nil {
		return nil, dberr
	}


	ids := make([]int64, len(followings))
	cacheMembers := make([]*models.CacheMember, len(followings))


	for i, f := range followings {
		ids[i] = f.FollowingID
		cacheMembers[i] = &models.CacheMember{
			Member: f.FollowingID,
			Score: makeScore(f.CreatedAt),
		}
	}

	tasks := cacheMembers
	err = r.pool.Submit(func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		_ = r.followCache.AddFollowings(bgCtx, userID, tasks)
	})

	if err != nil {
		slog.Warn("ants pool へのタスク投入に失敗しました。同期的なFollowingキャッシュ破棄を実行します。", "err", err)
		_ = r.followCache.InvalidateSelf(context.Background(), userID, true, false)
	}
		
	
	return ids, nil
}

func(r *followRepository) GetFollowers(ctx context.Context, userID int64) ([]int64, error) {
	list, err := r.followCache.GetFollowerIDs(ctx, userID)
    if err == nil && len(list) > 0 {
        return list, nil
    }

	sfKey := fmt.Sprintf("followers:%d", userID)

	followers, dberr := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func(c context.Context) ([]*models.Follow, error) {
		return r.followStore.GetFollowers(context.Background(), userID)
	})

	if dberr != nil {
		return nil, dberr
	}


	ids := make([]int64, len(followers))
	cacheMembers := make([]*models.CacheMember, len(followers))

	tasks := cacheMembers
	for i, f := range followers {
		ids[i] = f.FollowerID
		cacheMembers[i] = &models.CacheMember{
			Member: f.FollowerID,
			Score: makeScore(f.CreatedAt),
		}
	}
	
	err = r.pool.Submit(func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		
		_ = r.followCache.AddFollowers(bgCtx, userID, tasks)
	})

	if err != nil {
		slog.Warn("ants pool へのタスク投入に失敗しました。同期的なFollowerキャッシュ破棄を実行します。", "err", err)
		_ = r.followCache.InvalidateSelf(context.Background(), userID, false, true)
	}
	
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