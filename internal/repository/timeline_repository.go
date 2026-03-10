package repository

import (
	"aita/internal/dto"
	"aita/internal/models"
	"context"
	"time"

	"github.com/panjf2000/ants/v2"
)

type TimeLineCache interface {
	PushBatch(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error
	FindRange(ctx context.Context, userID int64, start, stop int64) ([]int64, error)
	RecallTweet(ctx context.Context, tweetID int64, userIDs []int64) error
	BackfillIDs(ctx context.Context, userID int64, tweets []*models.Tweet) error
}




type timeLineRepository struct {
	timeLineCache TimeLineCache
	pool       	  *ants.Pool
}

func NewTimeLineRepository(c TimeLineCache, p *ants.Pool) *timeLineRepository {
	return &timeLineRepository{
		timeLineCache: c,
		pool: p,
	}
}

func (r *timeLineRepository) Push(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error {
	err := r.timeLineCache.PushBatch(ctx, tweetID, userIDs, createdAt)
	return  err
}

func (r *timeLineRepository) GetHomeTimeLine(ctx context.Context, userID int64, page, size int) ([]int64, error) {
	start := int64(page*size)
	stop := start + int64(size) - 1

	return r.timeLineCache.FindRange(ctx, userID, start, stop)
	
}


func (r *timeLineRepository) Recall(ctx context.Context, tweetID int64, userIDs []int64) error {
	err := r.timeLineCache.RecallTweet(ctx, tweetID, userIDs)
	return err
}

func (r *timeLineRepository) Backfill(ctx context.Context, userID int64, records []*dto.TweetRecord) error {
	if len(records) == 0 {
		return nil
	}
	tweets := make([]*models.Tweet, len(records))
    for i, rec := range records {
        tweets[i] = rec.ToModel()
    }
    return r.timeLineCache.BackfillIDs(ctx, userID, tweets)
}

