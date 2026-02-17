package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresFollowStore struct {
	database *sqlx.DB
}

func NewPostgresFollowStore(db *sqlx.DB) *postgresFollowStore {
	return &postgresFollowStore{database: db}
}

func (s *postgresFollowStore)Create(ctx context.Context, follow *models.Follow) (*models.Follow, error) {
	query := `INSERT INTO follows(follower_id, following_id)
			  VALUES($1, $2)
			  RETURNING id, follower_id, following_id, created_at`
	var newFollow models.Follow
	err := s.database.QueryRowContext(
		ctx, 
		query,
		follow.FollowerID,
		follow.FollowingID, 
	).Scan(
		&newFollow.ID,
		&newFollow.FollowerID,
		&newFollow.FollowingID,
		&newFollow.CreatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
				case errCodeUniqueViolation:
    				if pqErr.Constraint == constraintUniqueFollow {
						return nil, errcode.ErrAlreadyFollowing 
					}
				case errCodeCheckViolation:
    				if pqErr.Constraint == constraintNoSelfFollow {
						return nil, errcode.ErrCannotFollowSelf
					}
				case errCodeForeignKeyViolation:
					return nil, errcode.ErrUserNotFound
			}
		}
		
		return nil,  fmt.Errorf("フォローの生成に失敗しました: %w", err)
	}

	return &newFollow, nil
}


func (s *postgresFollowStore) GetFollowing(ctx context.Context, followerID int64) ([]*models.Follow, error) {
	followings := []*models.Follow{}
	query := `SELECT id, follower_id, following_id, created_at FROM follows WHERE follower_id = $1`
	err := s.database.SelectContext(ctx, &followings, query, followerID)
	if err != nil {
		return nil, fmt.Errorf("フォロー中リストの取得に失敗しました: %w", err)
	}
	return followings, nil
}

func (s *postgresFollowStore) GetFollower(ctx context.Context, followingID int64) ([]*models.Follow, error) {
	followers := []*models.Follow{}
	query := `SELECT id, follower_id, following_id, created_at FROM follows WHERE following_id = $1`
	err := s.database.SelectContext(ctx, &followers, query, followingID)
	if err != nil {
		return nil, fmt.Errorf("フォロワーリストの取得に失敗しました: %w", err)
	}
	return followers, nil
}

func (s *postgresFollowStore) GetRelationship(ctx context.Context, userA, userB int64) (*models.RelationShip, error) {
    var relationship models.RelationShip
	query := `
        SELECT 
            EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2) AS following,
            EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = $1) AS followed_by
    `
	err := s.database.GetContext(ctx, &relationship, query, userA, userB)
    if err != nil {
        return nil, fmt.Errorf("関係性の取得に失敗しました: %w", err)
    }
    
    return &relationship, nil
}

func (s *postgresFollowStore) Delete(ctx context.Context, followerID, followingID int64) error {
    query := `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`
    _, err := s.database.ExecContext(ctx, query, followerID, followingID)
    if err != nil {
        return fmt.Errorf("フォロー解除に失敗しました: %w", err)
    }
    return nil
}




