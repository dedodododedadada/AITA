package service

import (
	"aita/internal/models"
	"context"
	"fmt"
)

type TweetStore interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
}

type tweetService struct {
	tweetStore TweetStore
}

func NewTweetService(ts TweetStore) *tweetService {
	return &tweetService{tweetStore: ts}
}

func (s *tweetService) PostTweet(ctx context.Context, userID int64, req *models.CreateTweetRequest) (*models.Tweet, error) {
	if userID <= 0 {
		return nil, models.ErrInvalidUserID
	}
	if err := models.IsValidTweetReq(req); err != nil {
		return nil, err
	}
	initialTweet := &models.Tweet{
		UserID:   userID,
		Content:  req.Content,
		ImageURL: req.ImageURL,
	}
	savedTweet, err := s.tweetStore.CreateTweet(ctx, initialTweet)
	if err != nil {
		return nil, fmt.Errorf("ツイートの挿入に失敗しました: %w", err)
	}
	return savedTweet, nil

}
