package repository

import (
	"aita/internal/dto"
	"context"
	"time"

	"github.com/panjf2000/ants/v2"
)

type TimeLineCache interface {
	PushBatch(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error
	FindRange(ctx context.Context, userID int64, start, stop int64) ([]int64, error)
	RecallTweet(ctx context.Context, tweetID int64, userIDs []int64) error
}

type TweetProvider interface {
	MultiGet(ctx context.Context, tweetIDs []int64) ([]*dto.TweetRecord, error)
}


type timeLineRepository struct {
	timeLineCache TimeLineCache
	tweetProvider TweetProvider
	pool       *ants.Pool
}

func NewTimeLineRepository(c TimeLineCache, t TweetProvider, p *ants.Pool) *timeLineRepository {
	return &timeLineRepository{
		timeLineCache: c,
		tweetProvider: t,
		pool: p,
	}
}

func (r *timeLineRepository) Push(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error {
	err := r.timeLineCache.PushBatch(ctx, tweetID, userIDs, createdAt)
	return  err
}

func (r *timeLineRepository) GetHomeTimeLine(ctx context.Context, userID int64, page, size int) ([]*dto.TweetRecord, error) {
	start := int64(page*size)
	stop := start + int64(size) - 1

	ids, err := r.timeLineCache.FindRange(ctx, userID, start, stop)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*dto.TweetRecord{}, nil
	}

	results, err  := r.tweetProvider.MultiGet(ctx, ids)

	if err != nil {
		return nil, err
	}
	
	return results, nil
}


func (r *timeLineRepository) Recall(ctx context.Context, tweetID int64, userIDs []int64) error {
	err := r.timeLineCache.RecallTweet(ctx, tweetID, userIDs)
	return err
}
