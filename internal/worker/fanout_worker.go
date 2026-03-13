package worker

import (
	"aita/internal/dto"
	"aita/internal/pkg/messagequeue"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type MQConsumer interface {
	Dequeue(ctx context.Context) (*messagequeue.MQMessage, error)
	Ack(ctx context.Context, msgID string) error
}

type FollwerProvider interface {
	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type TLHelper interface {
	Fanout(ctx context.Context, tweetID int64, targetIDs []int64,  createdAt time.Time) error 
	Forward(ctx context.Context, tweetID int64, userIDs []int64) error
}



type fanoutWorker struct {
	mQConsumer 			MQConsumer
	followerProvider 	FollwerProvider
	tLHelper   			TLHelper
	pool                *ants.Pool
}

func NewFanoutWorker(c MQConsumer, p FollwerProvider, h TLHelper, ap *ants.Pool) *fanoutWorker{
	return &fanoutWorker{
		mQConsumer: c,
		followerProvider: p,
		tLHelper: h,
		pool: ap,
	}
}

func (w *fanoutWorker) Start(ctx context.Context) {
    failCount := 0
    var circuitBreakerUntil time.Time 

    for {
        select {
        case <-ctx.Done():
            return
        default:
            if time.Now().Before(circuitBreakerUntil) {
                time.Sleep(5 * time.Second) 
                continue
            }

            message, err := w.mQConsumer.Dequeue(ctx)
            if err != nil {
                slog.Error("FanoutWorker: インフラ接続エラー", "error", err)
                time.Sleep(2 * time.Second)
                continue
            }
            if message == nil {
                continue
            }

            if err := w.handleTask(ctx, message.ID, message.Values); err == nil {
                failCount = 0
                _ = w.mQConsumer.Ack(ctx, message.ID)
            } else {
                failCount++
                slog.Error("FanoutWorker: タスク処理失敗", "count", failCount, "err", err)

                if failCount >= 20 {
                    slog.Warn("FanoutWorker: 連続失敗によりサーキットブレーカー発動。30秒間停止します。")
                    circuitBreakerUntil = time.Now().Add(30 * time.Second)
                    failCount = 0
                }
            }
        }
    }
}

func(w *fanoutWorker) handleTask(ctx context.Context, messageID string, values map[string]any) error {
	task :=&dto.FanoutTask{}
	err := task.FromMap(messageID, values)
	if err != nil {
		slog.Error("FanoutWorker: データ解析エラー。このメッセージを破棄します。", 
            "msg_id", messageID, 
            "error", err)
		return nil 
	}

	var bizErr error
    switch task.Action {
    case dto.ActionCreate:
        bizErr = w.processCreate(ctx, task)
    case dto.ActionDelete:
        bizErr = w.processDelete(ctx, task)
    default:
        slog.Warn("FanoutWorker: 未知のアクション", "action", task.Action)
        return nil 
    }

    return bizErr
}

func (w *fanoutWorker) processCreate(ctx context.Context, task *dto.FanoutTask) error {
    followers, err := w.followerProvider.GetFollowerIDs(ctx, task.AuthorID)
    if err != nil {
        return err
    }
    if len(followers) == 0 {
        return nil
    }

    var wg sync.WaitGroup
    var bizErr error
    var mu sync.Mutex 
	
	wg.Add(1)
    
    err = w.pool.Submit(func() {
        defer wg.Done()
        if err := w.tLHelper.Fanout(ctx, task.TweetID, followers, task.CreatedAt); err != nil {
            slog.Error("FanoutWorker: 拡散処理に失敗しました", "tweet_id", task.TweetID, "error", err)
            mu.Lock()
            bizErr = err 
            mu.Unlock()
        }
    })

    if err != nil {
        return err 
    }

    wg.Wait() 
    return bizErr 
}

func (w *fanoutWorker) processDelete(ctx context.Context, task *dto.FanoutTask) error {
    followers, err := w.followerProvider.GetFollowerIDs(ctx, task.AuthorID)
    if err != nil {
        return err
    }
    if len(followers) == 0 {
        return nil
    }

    var wg sync.WaitGroup
    var bizErr error
	var mu sync.Mutex 
    
    wg.Add(1)
    err = w.pool.Submit(func() {
        defer wg.Done()
        if err := w.tLHelper.Forward(ctx, task.TweetID, followers); err != nil {
            slog.Error("FanoutWorker: 削除(Forward)処理に失敗しました", 
                "tweet_id", task.TweetID, 
                "error", err,
            )
			mu.Lock()
            bizErr = err
			mu.Unlock()
        }
    })

    if err != nil {
        return err
    }

    wg.Wait()
    return bizErr
}

