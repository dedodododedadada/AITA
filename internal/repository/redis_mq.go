package repository

import (
	"aita/internal/dto"
	"aita/internal/pkg/utils"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisMQ struct {
	client 		*redis.Client
	stream 		string
	group  		string
	consumer 	string
}

func NewRedisMQ(c *redis.Client, stream, group, consumer string) *redisMQ {
	return &redisMQ{
		client: c,
		stream: stream,
		group: group,
		consumer: consumer,
	}
}

func (m *redisMQ) InitMQ(ctx context.Context) error {
    err := m.client.XGroupCreateMkStream(ctx, m.stream, m.group, "0").Err()
    if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
        return err
    }
    return nil
}

func (m *redisMQ) Dequeue(ctx context.Context) (*dto.FanoutTask, error) {
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
	tweetIDStr, _ := msg.Values["tweet_id"].(string) 
	authorIDStr, _ := msg.Values["author_id"].(string)
	return &dto.FanoutTask{
		MsgID: msg.ID,
		TweetID: utils.ParseInt64(tweetIDStr),
		AuthorID: utils.ParseInt64(authorIDStr),
	}, nil
}

func (m *redisMQ) ACK(ctx context.Context, msgID string) error {
	err := m.client.XAck(ctx, m.stream, m.group, msgID).Err()
	if err != nil {
		return err
	}
	return nil
}