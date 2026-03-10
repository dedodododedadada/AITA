package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type TweetRepository interface {
	Create(ctx context.Context, record *dto.TweetRecord) (*dto.TweetRecord, error) 
	Get(ctx context.Context, tweetID int64) (*dto.TweetRecord, error) 
	Update(ctx context.Context, newContent string, tweetID int64) (*dto.TweetRecord, error) 
	Delete(ctx context.Context, tweetID int64) error 
	MultiGet(ctx context.Context, tweetIDs []int64) ([]*dto.TweetRecord, error)
	GetTweetsByAuthor(ctx context.Context, userID int64, page, size int) ([]int64, error)
	AsyncToMQ(ctx context.Context, tweetID, authorID int64, createdAt time.Time, action string) error
}

type tweetService struct {
	tweetRepository TweetRepository
}

func NewTweetService(tr TweetRepository) *tweetService {
	return &tweetService{tweetRepository: tr}
}

func (s *tweetService) PostTweet(ctx context.Context, userID int64, content string, imageURL *string) (*dto.TweetRecord, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}

	if content == "" {
        return nil, errcode.ErrRequiredFieldMissing
    }

	initialTweet := &dto.TweetRecord{
		UserID:   userID,
		Content:  content,
		ImageURL: imageURL,
	}
	
	savedTweet, err := s.tweetRepository.Create(ctx, initialTweet)
	if err != nil {
		return nil, fmt.Errorf("ツイートの挿入に失敗しました: %w", err)
	}

	_ = s.tweetRepository.AsyncToMQ(
		ctx,
		savedTweet.ID,
		savedTweet.UserID,
		savedTweet.CreatedAt,
		dto.ActionCreate,
	)

	return savedTweet, nil
}

func (s *tweetService) FetchTweet(ctx context.Context, tweetID int64) (*dto.TweetRecord, error) {
	if tweetID <= 0 {
		return nil, errcode.ErrInvalidTweetID
	}

	tweet, err := s.tweetRepository.Get(ctx,  tweetID)

	if err != nil {
		return nil, fmt.Errorf("ツイート情報の取得に失敗しました: %w", err)
	}

	return tweet, nil
}

func (s *tweetService) ToMyTweet(ctx context.Context, tweetID int64, userID int64) (*dto.TweetRecord, error) {
	if userID <= 0 {
		return nil, errcode.ErrInvalidUserID
	}

	tweet, err := s.FetchTweet(ctx, tweetID)
	if err != nil {
		return nil, err
	}

	if tweet.UserID != userID {
		slog.Info("権限外のアクセスですが、メタデータは保持します", "tweet_id", tweet.ID)
		return tweet, errcode.ErrForbidden
	}

	return tweet, nil
}


func (s *tweetService) EditTweet(ctx context.Context, newContent string, tweetID int64, userID int64) (*dto.TweetRecord, bool, error) {
	if userID <= 0 {
		return nil, false, errcode.ErrInvalidUserID
	}

	if  newContent == "" {
		return nil, false, errcode.ErrRequiredFieldMissing
	}
	tweet, err := s.ToMyTweet(ctx, tweetID, userID)
	if err != nil {
		if err == errcode.ErrForbidden{
			return tweet, true, nil
		}
		return nil, false, err
	}

	if tweet.Content == newContent {
		return tweet, false, nil
	}

	if !tweet.CanBeUpdated() {
		return nil, false, errcode.ErrEditTimeExpired
	}

	tweet, err = s.tweetRepository.Update(ctx, newContent, tweetID)

	if err != nil {
		return nil, false, fmt.Errorf("ツイート編集に失敗しました: %w", err)
	}

	return tweet, true, nil
}

func (s *tweetService) RemoveTweet(ctx context.Context, tweetID int64, userID int64) error {
	if userID <= 0 {
		return errcode.ErrInvalidUserID
	}
	deletedTweet, err := s.ToMyTweet(ctx, tweetID, userID)
	if err != nil {
		if err == errcode.ErrForbidden {
			return nil
		}

		return err
	}

	err = s.tweetRepository.Delete(ctx, tweetID)
	if err != nil {
		return fmt.Errorf("ツイートの削除に失敗しました: %w", err)
	}

		_ = s.tweetRepository.AsyncToMQ(
		ctx,
		deletedTweet.ID,
		deletedTweet.UserID,
		deletedTweet.CreatedAt,
		dto.ActionDelete,
	)


	return nil
}


func (s *tweetService) GetTweets(ctx context.Context, tweetIDs []int64) ([]*dto.TweetRecord, error) {
    if len(tweetIDs) == 0 {
        return []*dto.TweetRecord{}, nil
    }

    tweets, err := s.tweetRepository.MultiGet(ctx, tweetIDs)
    if err != nil {
        return nil, fmt.Errorf("TimeLineService.GetTweets: ツイートリストの一括取得に失敗しました: %w", err)
    }

    return tweets, nil
}

func (s *tweetService) GetMyTweets(ctx context.Context, userID int64, page, size int) ([]*dto.TweetRecord, error) {
    if userID <= 0 {
        return nil, errcode.ErrInvalidUserID
    }
	
	if page < 0 { 
		page = 0
	}
    if size <= 0 || size > 100 { 
		size = 20 
	}

    ids, err := s.tweetRepository.GetTweetsByAuthor(ctx, userID, page, size)
    if err != nil {
        return nil, fmt.Errorf("TimeLineService.GetMyTweets: 投稿一覧の ID 取得に失敗しました (user_id: %d): %w", userID, err)
    }

    if len(ids) == 0 {
        return []*dto.TweetRecord{}, nil
    }

    tweets, err := s.tweetRepository.MultiGet(ctx, ids)
    if err != nil {
        return nil, fmt.Errorf("TimeLineService.GetMyTweets: 投稿内容のバルク変換に失敗しました (user_id: %d, count: %d): %w", 
            userID, len(ids), err)
    }

    return tweets, nil
}