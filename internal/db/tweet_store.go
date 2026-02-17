package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresTweetStore struct {
	database *sqlx.DB
}

func NewPostgresTweetStore(db *sqlx.DB) *postgresTweetStore {
	return &postgresTweetStore{database: db}
}

func (s *postgresTweetStore) CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error) {
	query := `
		INSERT INTO tweets(user_id, content, image_url)
		VALUES($1, $2, $3)
		RETURNING id, user_id, content, image_url, created_at, updated_at, is_edited`
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
		&newTweet.UpdatedAt,
		&newTweet.IsEdited, 
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == errCodeForeignKeyViolation && pqErr.Constraint == constraintTweetUserFK {
				return nil, errcode.ErrUserNotFound
			}
			if pqErr.Code == errCodeStringDataRightTruncation {
				return nil, errcode.ErrValueTooLong
			}
		}

		return nil, fmt.Errorf("ツイートの挿入に失敗しました: %w", err)
	}

	newTweet.CreatedAt = newTweet.CreatedAt.UTC()
	newTweet.UpdatedAt = newTweet.UpdatedAt.UTC()
	return &newTweet, nil
}

func (s *postgresTweetStore) GetTweetByTweetID(ctx context.Context, tweetID int64) (*models.Tweet, error) {
	query := `SELECT user_id, content, image_url, created_at, updated_at, is_edited FROM tweets WHERE id = $1`
	var wantedTweet models.Tweet
	err := s.database.GetContext(ctx, &wantedTweet, query, tweetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.ErrTweetNotFound
		}
		return nil, fmt.Errorf("ツイートの取得に失敗しました: %w", err)
	}
	wantedTweet.CreatedAt = wantedTweet.CreatedAt.UTC()
	wantedTweet.UpdatedAt = wantedTweet.UpdatedAt.UTC()
	return &wantedTweet, nil
}

func (s *postgresTweetStore) UpdateContent(ctx context.Context, newContent string, tweetID int64) (*models.Tweet, error) {
	query := `UPDATE tweets 
		SET content = $1, is_edited = true
		WHERE id = $2 AND content <> $1 
		RETURNING id, user_id, content, image_url, created_at, updated_at, is_edited`
	var updatedTweet models.Tweet
	err := s.database.GetContext(ctx, &updatedTweet, query, newContent, tweetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.ErrTweetNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == errCodeStringDataRightTruncation {
				return nil, errcode.ErrValueTooLong
			}
		}
		return nil, fmt.Errorf("ツイートの更新に失敗しました: %w", err)
	}

	updatedTweet.CreatedAt = updatedTweet.CreatedAt.UTC()
	updatedTweet.UpdatedAt = updatedTweet.UpdatedAt.UTC()
	return &updatedTweet, nil
}

func (s *postgresTweetStore) DeleteTweet(ctx context.Context, tweetID int64) error {
	query := `DELETE FROM tweets WHERE id = $1`
	result, err := s.database.ExecContext(ctx, query, tweetID)
	if err != nil {
		return fmt.Errorf("ツイートの削除に失敗しました: %w", err)
	}

	row, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗しました: %w", err)
	}

	if row == 0 {
		return errcode.ErrTweetNotFound
	}

	return nil
}
