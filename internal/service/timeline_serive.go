package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"context"
	"fmt"
	"time"
)

type TimeLineRepository interface {
	Push(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error
	GetHomeTimeLine(ctx context.Context, userID int64, page, size int) ([]*dto.TweetRecord, error) 
	Recall(ctx context.Context, tweetID int64, userIDs []int64) error 
}

type FollowerManager interface{
	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type timeLineService struct {
	timeLineRepository TimeLineRepository
	followerManager    FollowerManager
} 

func NewTimeLineService(r TimeLineRepository) *timeLineService {
	return &timeLineService{timeLineRepository: r}
}

func (s *timeLineService) Fanout(ctx context.Context, tweetID int64, authorID int64, createdAt time.Time) error {
	if tweetID <= 0 {
		return errcode.ErrInvalidTweetID
	}

	if authorID <= 0 {
		return errcode.ErrInvalidUserID
	}

	followerIDs, err := s.followerManager.GetFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}
	if len(followerIDs) == 0 {
		return nil
	}

	err = s.timeLineRepository.Push(ctx, tweetID, followerIDs, createdAt) 
	if err != nil {
		return fmt.Errorf("Fanout: タイムラインへのプッシュに失敗しました(tweet:%d, author:%d): %w", tweetID, authorID, err)
	}

	return nil
}

