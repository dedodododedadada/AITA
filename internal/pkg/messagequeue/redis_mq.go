package messagequeue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMQ struct {
	client 		*redis.Client
	stream 		string
	group  		string
	consumer 	string
}

type MQMessage struct {
    ID     string
    Values map[string]any
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


func (m *RedisMQ) Dequeue(ctx context.Context) (*MQMessage, error) {
	entries, err := m.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: m.group,
		Consumer: m.consumer,
		Streams: []string{m.stream, ">"},
		Count: 1,
		Block: 5*time.Second,
	}).Result()

	if err == redis.Nil || (err == nil && len(entries) == 0){
		return nil, nil
	}
	if err != nil {
		slog.Error("RedisMQ: Dequeue 失败 (ネットワークまたはRedisエラー)", "error", err)
		return nil, err
	}
	
	raw := entries[0].Messages[0]
		return &MQMessage{
		ID: raw.ID,
		Values: raw.Values,
	}, nil
}

func (m *RedisMQ) Ack(ctx context.Context, msgID string) error {
	err := m.client.XAck(ctx, m.stream, m.group, msgID).Err()
	if err != nil {
		return fmt.Errorf("RedisMQ.ACK: メッセージの確認応答に失敗しました (msg_id: %s, group: %s): %w", msgID, m.group, err)
	}
	return nil
}