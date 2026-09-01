package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	sf "aita/internal/pkg/singleflight"
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/panjf2000/ants/v2"
	"golang.org/x/sync/singleflight"
)

type TimeLineRepository interface {
	Push(ctx context.Context, tweetID int64, userIDs []int64, createdAt time.Time) error
	GetHomeTimeLine(ctx context.Context, userID int64, page, size int) ([]int64, error) 
	Recall(ctx context.Context, tweetID int64, userIDs []int64) error 
	Backfill(ctx context.Context, userID int64, tweets []*dto.TweetRecord) error
}

type TweetProvider interface  {
	GetTweets(ctx context.Context, tweetIDs []int64) ([]*dto.TweetRecord, error)
	GetMyTweets(ctx context.Context, userID int64, page, size int) ([]*dto.TweetRecord, error)
	GetRecentTweets(ctx context.Context, authorIDs []int64, page, size int) ([]*dto.TweetRecord, error)
}

type FollowProvider interface {
	GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error)
}

type timeLineService struct {
	timeLineRepository TimeLineRepository
	tweetProvider      TweetProvider
	followProvider     FollowProvider
	sf                 *singleflight.Group
	pool               *ants.Pool
} 

func NewTimeLineService(r TimeLineRepository, t TweetProvider, f FollowProvider, p *ants.Pool) *timeLineService {
	return &timeLineService{
		timeLineRepository: r,
		tweetProvider: t,
		followProvider: f,
		sf: &singleflight.Group{},
		pool: p,
	}
}

func (s *timeLineService) Fanout(ctx context.Context, tweetID int64, targetIDs []int64,  createdAt time.Time) error {
	if tweetID <= 0 {
		return errcode.ErrTweetNotFound
	}
	
	if len(targetIDs) == 0 {
		return nil
	}

	err := s.timeLineRepository.Push(ctx, tweetID, targetIDs, createdAt) 
	if err != nil {
		return fmt.Errorf("Fanout: タイムラインへのプッシュに失敗しました(tweet:%d): %w", tweetID, err)
	}

	return nil
}


func (s *timeLineService) Forward(ctx context.Context, tweetID int64, userIDs []int64) error {
	if tweetID <= 0 {
		return errcode.ErrTweetNotFound
	}
	if len(userIDs) == 0 {
		return nil
	}
	err := s.timeLineRepository.Recall(ctx, tweetID, userIDs)
	if err != nil {
		return fmt.Errorf("TimeLineService.Forward: タイムラインからの削除に失敗しました (tweet_id: %d, user_count: %d): %w", 
            tweetID, len(userIDs), err)
	}
	return nil
}

func (s *timeLineService) Backfill(ctx context.Context, userID int64, tweets []*dto.TweetRecord) error {
	if len(tweets) == 0 {
		return nil
	}
	err :=  s.timeLineRepository.Backfill(ctx, userID, tweets)

	if err != nil {
		return fmt.Errorf("TimeLineService.Backfill: タイムラインのキャッシュ書き込みに失敗しました (user_id: %d, count: %d): %w", 
            userID, len(tweets), err)
	}

	return nil
}

func (s *timeLineService) GetHomeTimeLine(ctx context.Context, userID int64, page, size int) ([]*dto.TweetRecord, error) {
	if userID <= 0 {
		return nil, errcode.ErrUserNotFound
	}

	if page < 0 || size <= 0 {
		return []*dto.TweetRecord{}, nil
	}

	ids, err := s.timeLineRepository.GetHomeTimeLine(ctx, userID, page, size)
	if err != nil {
        slog.Error("TimeLineService.GetHomeTimeLine: Redis からの ID 取得に失敗", "user_id", userID, "err", err)
    }

	var additionalTweets []*dto.TweetRecord
	var sfErr error
	if err != nil || (page == 0 && len(ids) < (size/2)) {
		sfKey := fmt.Sprintf("rebuildTimeLine:%d", userID)
		additionalTweets, sfErr = sf.GetDataWithSF(ctx, s.sf, sfKey, func(innerCtx context.Context) ([]*dto.TweetRecord, error) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			followingIDs, err := s.followProvider.GetFollowingIDs(bgCtx, userID)
			if err != nil {
				return nil, fmt.Errorf("TimeLineService.Rebuild: フォローリストの取得に失敗: %w", err)
			}

			targetAuthorIDs := append(followingIDs, userID)

			rebuildLimit := size - len(ids)

			if rebuildLimit <= 0 {
				rebuildLimit = size
			}

			rebuildTweets, err := s.tweetProvider.GetRecentTweets(bgCtx, targetAuthorIDs, 0, rebuildLimit)

			if err != nil {
				return nil, fmt.Errorf("TimeLineService.Rebuild: ツイートの取得に失敗: %w", err)
			}

			if len(rebuildTweets) == 0 {
				return []*dto.TweetRecord{}, nil
			}

			asyncErr := s.pool.Submit(func() {
				innerCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()

				_ = s.timeLineRepository.Backfill(innerCtx, userID, rebuildTweets)
			})
			if asyncErr != nil {
				slog.Warn("TimeLineService.Backfill: 非同期タスクの投入に失敗", "user_id", userID, "err", asyncErr)
			}

			return rebuildTweets, nil
		})

		if sfErr != nil {
            slog.Error("TimeLineService.Rebuild: タイムラインの再構築に失敗", "user_id", userID, "err", sfErr)
        } 
	}

	records, err := s.tweetProvider.GetTweets(ctx, ids)

	if err != nil {
		return nil, err
	}

	seen := make(map[int64]struct{}, len(records) + len(additionalTweets))
	finalResults := make([]*dto.TweetRecord, 0, len(records) + len(additionalTweets))

	for _, t := range append(records, additionalTweets...) {
		if t != nil {
			if _, exists := seen[t.ID]; !exists {
				seen[t.ID] = struct{}{}
				finalResults = append(finalResults, t)
			}
		}
	}

	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].CreatedAt.After(finalResults[j].CreatedAt)
	})

	return finalResults, nil
}


