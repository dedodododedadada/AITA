package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	sf "aita/internal/pkg/singleflight"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
	"golang.org/x/sync/singleflight"
)

type TweetStore interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
	GetTweetByTweetID(ctx context.Context, tweetID int64) (*models.Tweet, error)
	UpdateContent(ctx context.Context, newContent string, tweetID int64) (*models.Tweet, error)
	DeleteTweet(ctx context.Context, tweetID int64) error
	GetTweetsByTweetIDs(ctx context.Context, tweetIDS []int64) ([]*models.Tweet, error)
	GetTweetIDsByAuthor(ctx context.Context, authorID int64, page, size int) ([]int64, error) 
}

type TweetCache interface {
	SetTweet(ctx context.Context, tweet *models.Tweet) error
	GetTweet(ctx context.Context, tweetID int64) (*models.Tweet, error)
	Invalidate(ctx context.Context, tweetID int64) error
	MultiGetTweets(ctx context.Context, tweetIDs []int64) (map[int64]*models.Tweet, error)
	MultiSetTweets(ctx context.Context, tweets []*models.Tweet) error
}

type tweetRepository struct {
	tweetStore TweetStore
	tweetCache TweetCache
	sfTweet    *singleflight.Group
	pool       *ants.Pool
}

func NewTweetRepository(ts TweetStore, tc TweetCache, p *ants.Pool) *tweetRepository {
	return &tweetRepository{
		tweetStore: ts,
		tweetCache: tc,
		sfTweet: &singleflight.Group{},
		pool:       p,
	}
}

func (r *tweetRepository) Create(ctx context.Context, record *dto.TweetRecord) (*dto.TweetRecord, error) {
	tweet := record.ToModel()

	dbTweet, err := r.tweetStore.CreateTweet(ctx, tweet)
	if err != nil {
		return nil, err
	}

	if dbTweet == nil {
		return nil, errcode.ErrInternal
	}

	taskData := dbTweet

	err = r.pool.Submit(func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		innerErr := r.tweetCache.SetTweet(bgCtx, taskData)
		if innerErr != nil {
			_ = r.tweetCache.Invalidate(bgCtx, taskData.ID)
		}

	})
	if err != nil {
		_ = r.tweetCache.Invalidate(context.Background(), taskData.ID)
		slog.Warn("ants pool へのタスク投入に失敗しました。同期的なtweetキャッシュ破棄を実行します。", "err", err)
	}

	return dto.NewTweetRecord(dbTweet), nil
}

func (r *tweetRepository) Update(ctx context.Context, newContent string, tweetID int64) (*dto.TweetRecord, error) {
	tweet, err := r.tweetStore.UpdateContent(ctx, newContent, tweetID)

	if err != nil {
		return nil, err
	}

	_ = r.tweetCache.Invalidate(ctx, tweetID)

	_ = r.pool.Submit(func() {
		time.Sleep(800 * time.Millisecond)

		_ = r.tweetCache.Invalidate(context.Background(), tweetID)
	})
	return dto.NewTweetRecord(tweet), nil
}

func (r *tweetRepository) Delete(ctx context.Context, tweetID int64) error {
	err := r.tweetStore.DeleteTweet(ctx, tweetID)

	if err != nil {
		return err
	}

	_ = r.tweetCache.Invalidate(ctx, tweetID)

	return nil
}

func (r *tweetRepository) Get(ctx context.Context, tweetID int64) (*dto.TweetRecord, error) {
	tweet, err := r.tweetCache.GetTweet(ctx, tweetID)

	if err == nil && tweet != nil {
		return dto.NewTweetRecord(tweet), nil
	}

	sfKey := fmt.Sprintf("tweet:%d", tweetID)
	tweet, err = sf.GetDataWithSF(ctx, r.sfTweet, sfKey, func(innerCtx context.Context) (*models.Tweet, error) {
		return r.tweetStore.GetTweetByTweetID(innerCtx, tweetID)
	})

	if err != nil {
		return nil, err
	}

	if tweet == nil {
		return nil, errcode.ErrTweetNotFound
	}

	err = r.pool.Submit(func() {
		backfillCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		innerErr := r.tweetCache.SetTweet(backfillCtx, tweet)
		if innerErr != nil {
			r.tweetCache.Invalidate(context.Background(), tweet.ID)
		}
	})
	if err != nil {
		slog.Warn("Tweetバックフィルのタスク投入に失敗しました",
			"tweet_id", tweetID,
			"err", err,
		)
	}

	return dto.NewTweetRecord(tweet), nil
}

func (r *tweetRepository) MultiGet(ctx context.Context, tweetIDs []int64) ([]*dto.TweetRecord, error) {
	if len(tweetIDs) == 0 {
		return []*dto.TweetRecord{}, nil
	}

	tweetsMap, _ := r.tweetCache.MultiGetTweets(ctx, tweetIDs)
	if tweetsMap == nil {
		tweetsMap = make(map[int64]*models.Tweet, len(tweetIDs))
	}

	missedTIDs := make([]int64, 0, len(tweetIDs))
	for _, id := range tweetIDs {
		if _, found := tweetsMap[id]; !found {
			missedTIDs = append(missedTIDs, id)
		}
	}

	if len(missedTIDs) > 0 {
		dbTweets, err := r.tweetStore.GetTweetsByTweetIDs(ctx, missedTIDs)
		if err != nil {
			return nil, err
		}

		for _, tweet := range dbTweets {
			temp := tweet
			tweetsMap[tweet.ID] = temp
		}

		err = r.pool.Submit(func() {
			backfillCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_ = r.tweetCache.MultiSetTweets(backfillCtx, dbTweets)
		})

		if err != nil {
			slog.Warn("Tweetリストの一括バックフィル投入に失敗しました",
				"missed_count", len(missedTIDs),
				"err", err,
			)
		}
	}

	finalResults := make([]*dto.TweetRecord, 0, len(tweetIDs))
	for _, id := range tweetIDs {
		if tweet, ok := tweetsMap[id]; ok {
			finalResults = append(finalResults, dto.NewTweetRecord(tweet))
		}
	}

	return finalResults, nil
}


func (r *tweetRepository) GetTweetsByAuthor(ctx context.Context, userID int64, page, size int) ([]int64, error) {
	if page < 0 || size <= 0 {
		return []int64{}, nil
	}

	sfKey := fmt.Sprintf("GetMytweets:%d:page:%d:size%d", userID, page, size)
	ids, err := sf.GetDataWithSF(ctx, r.sfTweet, sfKey, func(innerCtx context.Context) ([]int64, error) {
		return r.tweetStore.GetTweetIDsByAuthor(innerCtx, userID, page, size)
	})

	if err != nil {
		return nil, err
	}

	return ids, nil
}