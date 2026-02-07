package db

import (
	"aita/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresTweetStore struct {
	database *sqlx.DB
}

func NewPostgresTweetStore(DB *sqlx.DB) *postgresTweetStore {
	return &postgresTweetStore{database: DB}
}

func (s *postgresTweetStore) CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error) {
	query := `
		INSERT INTO tweets(user_id, content, image_url)
		VALUES($1, $2, $3)
		RETURNING id, user_id, content, image_url, created_at`
	var newTweet models.Tweet
	err := s.database.QueryRowContext(
		ctx,
		query,
		tweet.UserID,
		tweet.Content,
		tweet.ImageURL,
	).Scan(
		&newTweet.ID, 
		&newTweet.UserID,
		&newTweet.Content,
		&newTweet.ImageURL,
		&newTweet.CreatedAt,
	)
	
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == errCodeForeignKeyViolation && pqErr.Constraint == constraintTweetUserFK {
                return nil, models.ErrUserNotFound
            }
			if pqErr.Code == errCodeStringDataRightTruncation {
				return nil, models.ErrValueTooLong
			}
		}
		
		return nil, fmt.Errorf("ツイートの挿入に失敗しました: %w", err)
	}

	newTweet.CreatedAt = newTweet.CreatedAt.UTC()
	return &newTweet, nil
}

