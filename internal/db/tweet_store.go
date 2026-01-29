package db

import (
	"aita/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TweetStore interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet) (error)
}

type PostgresTweetStore struct {
	database *sqlx.DB
}

func NewPostgresTweetStore(DB *sqlx.DB) *PostgresTweetStore {
	return &PostgresTweetStore{database: DB}
}

func (s *PostgresTweetStore) CreateTweet(ctx context.Context, tweet *models.Tweet) error {
	query := `
		INSERT INTO tweets(user_id, content, image_url)
		VALUES($1, $2, $3)
		RETURNING id, created_at`
	err := s.database.QueryRowContext(
		ctx,
		query,
		tweet.UserID,
		tweet.Content,
		tweet.ImageURL,
	).Scan(&tweet.ID, &tweet.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr){
			if pqErr.Code == errCodeForeignKeyViolation && pqErr.Constraint == constraintTweetUserFK {
                return models.ErrUserNotFound
            }
		}
		
		return fmt.Errorf("ツイートの挿入に失敗しました: %w", err)
	}
	return nil
}

