package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"context"
	"fmt"
)

type FollowRepository interface {
	Create(ctx context.Context, followerID, followingID int64) (*dto.FollowRecord, error)
	CheckRelation(ctx context.Context, followerID, followingID int64) (*dto.RelationRecord, error) 
	GetFollowings(ctx context.Context, userID int64) ([]int64, error) 
	GetFollowers(ctx context.Context, userID int64) ([]int64, error)
	RemoveFollow(ctx context.Context, followerID, followingID int64) error 
}

type CountManager interface {
	UpdateFollowingCount(ctx context.Context, userID int64, delta int64) error
	UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error
	Exists(ctx context.Context, userID int64) (bool, error)
	GetInfoLists(ctx context.Context, userIDs []int64) ([]*dto.UserSlimRecord, error)
}

type TransactionManager interface {
	Exec(ctx context.Context, fn func(ctx context.Context) error) error
}

type followeService struct {
	followRepository 	FollowRepository
	countManager     	CountManager
	transactionManager  TransactionManager
}

func NewFollowService(fr FollowRepository, cm CountManager) *followeService {
	return &followeService{
		followRepository: fr,
		countManager: cm,
	}
}

func (s *followeService) Follow(ctx context.Context, userID, targetID int64) (*dto.FollowRecord, error) {
	if userID <= 0 || targetID <= 0{
		return nil, errcode.ErrInvalidUserID
	}

	if userID == targetID {
		return nil, errcode.ErrCannotFollowSelf
	}

	exists, err := s.countManager.Exists(ctx, targetID)
    if err != nil || !exists {
        return nil, errcode.ErrUserNotFound
    }

	var record *dto.FollowRecord
	err = s.transactionManager.Exec(ctx, func(txCtx context.Context) error {
        var err error
        record, err = s.followRepository.Create(txCtx, userID, targetID)
        if err != nil {
            return fmt.Errorf("followに失敗しました:%w", err) 
        }

        if err := s.countManager.UpdateFollowerCount(txCtx, targetID, 1); err != nil {
            return fmt.Errorf("フォロワー数の更新に失敗しました:%w", err) 
        }

        if err := s.countManager.UpdateFollowingCount(txCtx, userID, 1); err != nil {
            return fmt.Errorf("フォロー数の更新に失敗しました:%w", err) 
        }

        return nil
    })

    if err != nil {
        return nil, err 
    }

    return record, nil
}

func(s *followeService) UnFollow(ctx context.Context, userID, targetID int64) error {
	if userID <=0 || targetID <= 0 {
		return errcode.ErrInvalidUserID
	} 

	if userID == targetID {
		return errcode.ErrCannotFollowSelf
	}

	exists, err := s.countManager.Exists(ctx, targetID)
    if err != nil || !exists {
        return errcode.ErrUserNotFound
    }

	err = s.transactionManager.Exec(ctx, func(txCtx context.Context) error {
		err := s.followRepository.RemoveFollow(txCtx, userID, targetID)
		if err != nil {
			return fmt.Errorf("関係の削除に失敗しました:%w", err) 
		}

		err = s.countManager.UpdateFollowerCount(txCtx, targetID, -1)
		if err != nil {
			return fmt.Errorf("フォロワー数の減算に失敗しました:%w", err) 
		}

		err = s.countManager.UpdateFollowingCount(txCtx, userID, -1)
		if err != nil {
			return fmt.Errorf("フォロー数の減算に失敗しました:%w", err) 
		}

		return nil
	})

	if err != nil {
        return  err 
    }
	
	return nil
}

func(s *followeService) GetFollowers(ctx context.Context, userID int64) ([]*dto.UserSlimRecord, error) {
    if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}

	followerIDs, err := s.followRepository.GetFollowers(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("フォロワーの取得に失敗しました%w",err)
    }
	
	if len(followerIDs) == 0 {
		return []*dto.UserSlimRecord{}, nil
	}

    return s.countManager.GetInfoLists(ctx, followerIDs)
}

func(s *followeService) GetFollowings(ctx context.Context, userID int64) ([]*dto.UserSlimRecord, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}

	followingIDs, err := s.followRepository.GetFollowings(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("フォロー数の取得に失敗しました:%w", err)
    }

	if len(followingIDs) == 0 {
		return []*dto.UserSlimRecord{}, nil
	}

    return s.countManager.GetInfoLists(ctx, followingIDs)
}

func (s *followeService) GetRelation(ctx context.Context, userID, targetID int64) (*dto.RelationRecord, error) {
    if userID <= 0 || targetID <= 0 {
        return nil, errcode.ErrInvalidUserID
    }

    if userID == targetID {
        return nil, errcode.ErrCannotFollowSelf
    }


    record, err := s.followRepository.CheckRelation(ctx, userID, targetID)
    if err != nil {
        return nil, fmt.Errorf("GetRelation: 関係情報の取得に失敗しました(viewer:%d, target:%d): %w", userID, targetID, err)
    }

    return record, nil
}