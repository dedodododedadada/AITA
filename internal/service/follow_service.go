package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"context"
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
            return err 
        }

        if err := s.countManager.UpdateFollowerCount(txCtx, userID, 1); err != nil {
            return err
        }

        if err := s.countManager.UpdateFollowingCount(txCtx, targetID, 1); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, err 
    }

    return &dto.FollowRecord{
        FollowerID:    record.FollowerID,
        FollowingID:  record.FollowingID,
        CreatedAt: record.CreatedAt,
    }, nil
}