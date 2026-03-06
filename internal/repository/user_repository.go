package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
	"golang.org/x/sync/singleflight"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetFullByID(ctx context.Context, userID int64) (*models.User, error) 
	IncreaseFollowerCount(ctx context.Context,  userID, delta int64) error
	IncreaseFollowingCount(ctx context.Context, userID, delta int64) error
	GetNamesByIDs(ctx context.Context, userIDs []int64) ([]*models.UserInfo, error) 
}

type UserCache interface {
	Add(ctx context.Context, info *models.UserInfo, follower, following int64) error
	Invalidate(ctx context.Context, userID int64) 
	Get(ctx context.Context, userID int64) (*models.UserInfo, int64, int64, error) 
	Exists(ctx context.Context, userID int64) (bool, error)
	GetLists(ctx context.Context, userIDs []int64) (map[int64]*models.UserInfo, error)
	AddLists(ctx context.Context, infos []*models.UserInfo) error
	
}

type userRepository struct {
	userStore UserStore
	userCache UserCache
	sfUser  *singleflight.Group
	pool      *ants.Pool
}

func NewUserRepository(us UserStore, uc UserCache, p *ants.Pool) *userRepository {
	return &userRepository{
		userStore: us,
		userCache: uc,
		sfUser: &singleflight.Group{},
		pool: p, 
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
	res, err := utils.GetDataWithSF(ctx, r.sfUser, sfkey, func(c context.Context) (*models.User, error) {
		return r.userStore.GetByEmail(ctx, email)
	})

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, errcode.ErrUserNotFound
	}
	return dto.NewUserRecord(res), nil
}

func(r *userRepository) GetFullByID(ctx context.Context, userID int64) (*dto.UserRecord, error) {
	sfKey := fmt.Sprintf("users:%d", userID)

	user, err := utils.GetDataWithSF(ctx, r.sfUser, sfKey, func(c context.Context) (*models.User, error) {
		return r.userStore.GetFullByID(ctx, userID)
	})

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errcode.ErrUserNotFound
	}

	return dto.NewUserRecord(user), nil
}

func (r *userRepository) GetProfByID(ctx context.Context, userID int64) (*dto.UserPageRecord, error) {
    info, frc, fgc, err := r.userCache.Get(ctx, userID)
    if err == nil && info != nil {
        return dto.NewUserPageRecord(info, frc, fgc), nil
    }

    sfKey := fmt.Sprintf("prof:%d", userID)
    user, err := utils.GetDataWithSF(ctx, r.sfUser, sfKey, func(c context.Context) (*models.User, error) {
        return r.userStore.GetFullByID(ctx, userID) 
    })

    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errcode.ErrUserNotFound
    }

    err = r.pool.Submit(func() {
        backfillCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        err = r.userCache.Add(backfillCtx, user.ToCacheInfo(), user.FollowerCount, user.FollowingCount)
		if err != nil {
			r.userCache.Invalidate(context.Background(), user.ID)
		}
    })
	if err != nil {
		r.userCache.Invalidate(context.Background(), user.ID)
		slog.Warn("ants pool へのタスク投入失敗。プロフィールのバックフィルをスキップします。", "userID", userID, "err", err)
    }

    return &dto.UserPageRecord{
		ID: user.ID,
		Username: user.Username,
		FollowerCount: user.FollowerCount,
		FollowingCount: user.FollowingCount,

	}, nil
}

func (r *userRepository) IncreaseFollower(ctx context.Context, userID int64, delta int64) error {
    err := r.userStore.IncreaseFollowerCount(ctx,  userID, delta)
    if err != nil {
        return err 
    }
    
	r.userCache.Invalidate(ctx, userID)

    return nil
}

func (r *userRepository) IncreaseFollowing(ctx context.Context, userID int64, delta int64) error {
    err := r.userStore.IncreaseFollowingCount(ctx, userID, delta)
    if err != nil {
        return err 
    }

    r.userCache.Invalidate(ctx, userID)

    return nil
}

func (r *userRepository) Exists(ctx context.Context, id int64) (bool, error) {
    found, err := r.userCache.Exists(ctx, id)
    if err == nil && found {
        return true, nil
    }

    sfKey := fmt.Sprintf("exists:%d", id)
    exists, err := utils.GetDataWithSF(ctx, r.sfUser, sfKey, func(c context.Context) (bool, error) {
        _, dbErr := r.userStore.GetFullByID(ctx, id)
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

func (r *userRepository) GetBaseInfos(ctx context.Context, userIDs []int64) ([]*dto.UserSlimRecord, error) {
	if len(userIDs) == 0 {
		return []*dto.UserSlimRecord{}, nil
	}

	infos, _ := r.userCache.GetLists(ctx, userIDs)
	if infos == nil {
        infos = make(map[int64]*models.UserInfo, len(userIDs))
    }

	missedIDs := make([]int64, 0, len(userIDs))
	for _, id := range userIDs {
		if _, found := infos[id]; !found {
			missedIDs = append(missedIDs, id)
		}
	}

	if len(missedIDs) > 0  {
		dbInfos, err := r.userStore.GetNamesByIDs(ctx, missedIDs)
		if err != nil {
			return nil, err
		}

		for _, info := range dbInfos {
			temp := info
			infos[info.ID] = temp
		}

		err = r.pool.Submit(func() {
			backfillCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()

			_ = r.userCache.AddLists(backfillCtx, dbInfos)
		})

		if err != nil {
			slog.Warn("一括バックフィルがスキップされました", "count", len(missedIDs))
		}
	}

	finalResults := make([]*dto.UserSlimRecord, 0, len(userIDs))
    for _, id := range userIDs {
        if info, ok := infos[id]; ok {
            finalResults = append(finalResults, &dto.UserSlimRecord{
                ID:       info.ID,
                Username:  info.Username,
            })
        }
    }

	return finalResults, nil
}