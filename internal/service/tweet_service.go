package service

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"fmt"
)

type TweetStore interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
	GetTweetByTweetID(ctx context.Context, tweetID int64) (*models.Tweet, error)
	UpdateContent(ctx context.Context, newContent string, tweetID int64) (*models.Tweet,error)
	DeleteTweet(ctx context.Context, tweetID int64) error
}

type tweetService struct {
	tweetStore TweetStore
}

func NewTweetService(ts TweetStore) *tweetService {
	return &tweetService{tweetStore: ts}
}

func (s *tweetService) PostTweet(ctx context.Context, userID int64, content string, imageURL *string) (*models.Tweet, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}

	if content == "" {
        return nil, errcode.ErrRequiredFieldMissing
    }

	initialTweet := &models.Tweet{
		UserID:   userID,
		Content:  content,
		ImageURL: imageURL,
	}
	
	savedTweet, err := s.tweetStore.CreateTweet(ctx, initialTweet)
	if err != nil {
		return nil, fmt.Errorf("ツイートの挿入に失敗しました: %w", err)
	}
	return savedTweet, nil
}

func (s *tweetService) FetchTweet(ctx context.Context, tweetID int64) (*models.Tweet, error) {
	if tweetID <= 0 {
		return nil, errcode.ErrInvalidTweetID
	}

	tweet, err := s.tweetStore.GetTweetByTweetID(ctx, tweetID)

	if err != nil {
		return nil, fmt.Errorf("ツイート情報の取得に失敗しました: %w", err)
	}

	return tweet, nil
}

func (s *tweetService) ToMyTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}

	tweet, err := s.FetchTweet(ctx, tweetID)
	if err != nil {
		return nil, err
	}

	if tweet.UserID != userID {
		return nil, errcode.ErrForbidden
	}

	return tweet, nil
}

func (s *tweetService) EditTweet(ctx context.Context, newContent string, tweetID int64, userID int64) (*models.Tweet, bool, error) {
	if userID <= 0 {
		return nil, false, errcode.ErrInvalidUserID
	}

	if  newContent == "" {
		return nil, false, errcode.ErrRequiredFieldMissing
	}
	tweet, err := s.ToMyTweet(ctx, tweetID, userID)
	if err != nil {
		return nil, false, err
	}

	if tweet.Content == newContent {
		return tweet, false, nil
	}

	if !tweet.CanBeUpdated() {
		return nil, false, errcode.ErrEditTimeExpired
	}

	tweet, err = s.tweetStore.UpdateContent(ctx, newContent, tweetID)

	if err != nil {
		return nil, false, fmt.Errorf("ツイート編集に失敗しました: %w", err)
	}

	return tweet, true, nil
}

func (s *tweetService) RemoveTweet(ctx context.Context, tweetID int64, userID int64) error {
	if userID <= 0 {
		return errcode.ErrInvalidUserID
	}
	_, err := s.ToMyTweet(ctx, tweetID, userID)
	if err != nil {
		return err
	}

	err = s.tweetStore.DeleteTweet(ctx, tweetID)
	if err != nil {
		return fmt.Errorf("ツイートの削除に失敗しました: %w", err)
	}

	return nil
}
