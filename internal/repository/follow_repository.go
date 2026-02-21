package repository

import (
	"aita/internal/models"
	"context"
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
	RemoveFollowing(ctx context.Context, followerID, followingID int64) error
	Invalidate(ctx context.Context, userID int64) error
}
type FollowRepository struct {
	followStore FollowStore
	followCache FollowCache
}

func NewFollowRepository(fs FollowStore, fc FollowCache) *FollowRepository {
	return &FollowRepository{
		followStore: fs,
		followCache: fc,
	}
}

