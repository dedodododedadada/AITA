package service

import (
	"aita/internal/db"
	"aita/internal/models"
	"context"
	"fmt"
)

type TweetService struct {
	tweetStore db.TweetStore
}

func NewTweetService(ts  db.TweetStore) *TweetService {
	return &TweetService{tweetStore: ts}
}

func(s *TweetService) PostTweet(ctx context.Context, userID int64, req models.CreateTweetRequest) (*models.Tweet, error) {
	
	if req.Content == "" {
		return nil, models.ErrContentEmpty
	}

	tweet := &models.Tweet{
		UserID: userID,
		Content: req.Content,
		ImageURL: req.ImageURL,
	}

	if err := s.tweetStore.CreateTweet(ctx, tweet); err != nil {
		return nil, fmt.Errorf("failed to create tweet: %w", err)
	}

	return tweet, nil

}