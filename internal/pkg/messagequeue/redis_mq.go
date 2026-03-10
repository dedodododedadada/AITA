package messagequeue

import (
	"aita/internal/dto"
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMQ struct {
	client 		*redis.Client
	stream 		string
	group  		string
	consumer 	string
}

func NewRedisMQ(c *redis.Client, stream, group, consumer string) *RedisMQ {
	return &RedisMQ{
		client: c,
		stream: stream,
		group: group,
		consumer: consumer,
	}
}

func (m *RedisMQ) InitMQ(ctx context.Context) error {
    err := m.client.XGroupCreateMkStream(ctx, m.stream, m.group, "0").Err()
    if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
        return err
    }
    return nil
}

func (m *RedisMQ) Enqueue(ctx context.Context, values map[string]any) error {
    err :=  m.client.XAdd(ctx, &redis.XAddArgs{
        Stream: m.stream,
        MaxLen: 100000,
        Approx: true,
        Values: values,
    }).Err()

	if err != nil {
        return fmt.Errorf("RedisMQ.Enqueue: メッセージの投入失敗しました(%v):%w" ,values, err)
    }
	return nil
}


func (m *RedisMQ) Dequeue(ctx context.Context) (*dto.FanoutTask, error) {
	entries, err := m.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: m.group,
		Consumer: m.consumer,
		Streams: []string{m.stream, ">"},
		Count: 1,
		Block: 5*time.Second,
	}).Result()

	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	msg := entries[0].Messages[0]
	tweetIDStr, okT := msg.Values["tweet_id"].(string) 
	authorIDStr, okA := msg.Values["author_id"].(string)
	createdAtStr, okC := msg.Values["created_at"].(string) 
	if !okT || !okA || !okC {
        return nil, fmt.Errorf("RedisMQ.Dequeue: メッセージの型変換に失敗しました (msg_id: %s, raw_values: %v)", msg.ID, msg.Values)
    }

	return &dto.FanoutTask{
		MsgID: msg.ID,
		TweetID: utils.ParseInt64(tweetIDStr),
		AuthorID: utils.ParseInt64(authorIDStr),
		CreatedAt: time.Unix(utils.ParseInt64(createdAtStr), 0),
	}, nil
}

func (m *RedisMQ) ACK(ctx context.Context, msgID string) error {
	err := m.client.XAck(ctx, m.stream, m.group, msgID).Err()
	if err != nil {
		return fmt.Errorf("RedisMQ.ACK: メッセージの確認応答に失敗しました (msg_id: %s, group: %s): %w", msgID, m.group, err)
	}
	return nil
}