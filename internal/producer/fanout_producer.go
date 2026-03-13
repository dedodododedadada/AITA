package producer

import (
	"aita/internal/dto"
	"aita/internal/pkg/messagequeue"
	"context"
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
)

type Enqueueer interface {
	Enqueue(ctx context.Context, values map[string]any) error 
}

type fanoutProducer struct {
	enqueuer Enqueueer
	pool     *ants.Pool
}

func NewFanoutProducer(r *messagequeue.RedisMQ, p *ants.Pool) *fanoutProducer {
	return &fanoutProducer{
		enqueuer: r,
		pool: p,
	}

}

func (f *fanoutProducer) AsyncToMQ(ctx context.Context, tweetID, authorID int64, createdAt time.Time, action string) error {
	task := dto.NewFanoutTask(tweetID, authorID, createdAt, action)
	taskMap := task.ToMap()
	err := f.pool.Submit(func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		innerErr := f.enqueuer.Enqueue(bgCtx, taskMap)
		if innerErr != nil {
			slog.Error("MQへの非同期書き込みに失敗しました",
				"tweet_id", tweetID,
				"action", action,
				"err", innerErr,
			)
		}
	})
	if err != nil {
		slog.Warn("アンツプールへのタスク投入に失敗しました",
			"tweet_id", tweetID,
			"action", action,
			"err", err,
		)
	}

	return nil
}