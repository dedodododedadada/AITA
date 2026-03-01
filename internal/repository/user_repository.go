package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
)

type Userstore interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, userID int64) (*models.User, error) 
	IncreaseFollowerCount(ctx context.Context,  userID, delta int64) error
	IncreaseFollowingCount(ctx context.Context, userID, delta int64) error
}

type UserCache interface {
	Add(ctx context.Context, info *models.UserCacheInfo, follower, following int64)
	Invalidate(ctx context.Context, userID int64) 
	Get(ctx context.Context, userID int64) (*models.UserCacheInfo, int64, int64, error) 
	IncrFollower(ctx context.Context, userID int64, delta int64) error
	IncrFollowing(ctx context.Context, userID int64, delta int64) error 
	Exists(ctx context.Context, userID int64) (bool, error)
}

type userRepository struct {
	userStore Userstore
	userCache UserCache
	sfFollow  *singleflight.Group
}

func NewUserRepository(us Userstore, uc UserCache) *userRepository {
	return &userRepository{
		userStore: us,
		userCache: uc,
	}
}

func(r *userRepository) Create(ctx context.Context, record *dto.UserRecord) ( *dto.UserRecord, error) {
	user := record.ToUserModel()

	dbUser, err := r.userStore.Create(ctx, user)

	if err != nil {
		return nil, err
	}

	if dbUser == nil {
		return nil, errcode.ErrInternal
	}

	return dto.NewUserRecord(dbUser), nil
}

func(r *userRepository) GetByEmail(ctx context.Context, email string) (*dto.UserRecord, error) {
	sfkey := fmt.Sprintf("users:%s", email)
	res, err := utils.GetDataWithSF(ctx, r.sfFollow, sfkey, func() (*models.User, error) {
		return r.userStore.GetByEmail(context.Background(), email)
	})

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, errcode.ErrInternal
	}
	return dto.NewUserRecord(res), nil
}

func(r *userRepository) GetByID(ctx context.Context, userID int64) (*dto.UserRecord, error) {
	info, follower, following, err := r.userCache.Get(ctx, userID)
	if err == nil && info != nil {
		user := info.ToUser(follower, following)
		return dto.NewUserRecord(user), nil
	}

	sfKey := fmt.Sprintf("users:%d", userID)

	user, err := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func() (*models.User, error) {
		return r.userStore.GetByID(context.Background(), userID)
	})

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errcode.ErrInternal
	}
    
	go func() {
		 r.userCache.Add(context.Background(), user.ToCacheInfo(), user.FollowerCount, user.FollowingCount)
	}()

	return dto.NewUserRecord(user), nil
}

func (r *userRepository) IncreaseFollower(ctx context.Context, userID int64, delta int64) error {
    err := r.userStore.IncreaseFollowerCount(ctx,  userID, delta)
    if err != nil {
        return err 
    }

    exists, err := r.userCache.Exists(ctx, userID)
    if err != nil {
        return err
    }

    if exists {
        err = r.userCache.IncrFollower(ctx, userID, delta)
        if err != nil {
            r.userCache.Invalidate(ctx, userID)
        }
    }
    return nil
}

func (r *userRepository) IncreaseFollowing(ctx context.Context, userID int64, delta int64) error {
    err := r.userStore.IncreaseFollowingCount(ctx, userID, delta)
    if err != nil {
        return err 
    }

    exists, err := r.userCache.Exists(ctx, userID)
    if err != nil {
        return err
    }

    if exists {
        err = r.userCache.IncrFollowing(ctx, userID, delta)
        if err != nil {
            r.userCache.Invalidate(ctx, userID)
        }
    }
    return nil
}

func (r *userRepository) Exists(ctx context.Context, id int64) (bool, error) {
    found, err := r.userCache.Exists(ctx, id)
    if err == nil && found {
        return true, nil
    }

    sfKey := fmt.Sprintf("exists:%d", id)
    exists, err := utils.GetDataWithSF(ctx, r.sfFollow, sfKey, func() (bool, error) {
        _, dbErr := r.userStore.GetByID(context.Background(), id)
        if dbErr != nil {
            if errors.Is(dbErr, errcode.ErrUserNotFound) {
                return false, nil
            }
            return false, dbErr
        }
        return true, nil
    })
	return exists, err
}