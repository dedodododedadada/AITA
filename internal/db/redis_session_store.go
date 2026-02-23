package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisSessionStore struct {
	client *redis.Client
	prefix string
}

func NewRedisSessionStore(client *redis.Client) *redisSessionStore {
	return &redisSessionStore{
		client: client,
		prefix: "session",
	}
}

func (s *redisSessionStore) dataKey(hash string) string {
	return fmt.Sprintf("%sdata:%s", s.prefix, hash)
}

func (s *redisSessionStore) hashKey(userID int64) string {
	return fmt.Sprintf("%shash:%d", s.prefix, userID)
}

func (s *redisSessionStore) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	dKey := s.dataKey(session.TokenHash)
	hKey := s.hashKey(session.UserID)

	data, err := json.Marshal(session)

	if err != nil {
		return nil, fmt.Errorf("セッションのシリアライズに失敗しました: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return nil, errcode.ErrSessionExpired
	}
	pipe := s.client.Pipeline()

	pipe.Set(ctx, dKey, data, ttl)
	pipe.SAdd(ctx, hKey, session.TokenHash)
	pipe.Expire(ctx, hKey, ttl+24*time.Hour)

	_, err = pipe.Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("Redisへのセッション保存に失敗しました: %w", err)
	}

	return session, nil
}

func (s *redisSessionStore) Get(ctx context.Context, tokenHash string) (*models.Session, error) {
	dKey := s.dataKey(tokenHash)

	val, err := s.client.Get(ctx, dKey).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errcode.ErrSessionNotFound
		}

		return nil, fmt.Errorf("Redisからの取得に失敗しました: %w", err)
	}

	var newSession models.Session

	if err := json.Unmarshal([]byte(val), &newSession); err != nil {
		return nil, fmt.Errorf("セッションのデシリアライズに失敗しました: %w", err)
	}

	if newSession.ExpiresAt.Before(time.Now()) {
		return nil, errcode.ErrSessionExpired
	}

	return &newSession, nil
}

func (s *redisSessionStore) Update(ctx context.Context, session *models.Session) error {
	dKey := s.dataKey(session.TokenHash)
	hKey := s.hashKey(session.UserID)
	data, err := json.Marshal(session)

	if err != nil {
		return fmt.Errorf("セッションのシリアライズに失敗しました: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errcode.ErrSessionExpired
	}

	pipe := s.client.Pipeline()

	pipe.Set(ctx, dKey, data, ttl)
	pipe.SAdd(ctx, hKey, session.TokenHash)
	pipe.Expire(ctx, hKey, ttl+24*time.Hour)

	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("Redisへのセッション更新に失敗しました: %w", err)
	}
	return nil
}

func (s *redisSessionStore) Delete(ctx context.Context, session *models.Session) error {
	dKey := s.dataKey(session.TokenHash)
    uKey := s.hashKey(session.UserID) 

    pipe := s.client.Pipeline()

    pipe.Del(ctx, dKey)
    pipe.SRem(ctx, uKey, session.TokenHash)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("セッションの完全削除に失敗しました: %w", err)
    }

	return nil
}

func (s *redisSessionStore) DeleteByUserID(ctx context.Context, userID int64) error {
    uKey := s.hashKey(userID)

    hashes, err := s.client.SMembers(ctx, uKey).Result()
    if err != nil {
        return fmt.Errorf("インデックスの取得に失敗しました: %w", err)
    }
    if len(hashes) == 0 {
        return nil
    }

    pipe := s.client.Pipeline()

    for _, hash := range hashes {
        sKey := s.dataKey(hash)
        pipe.Del(ctx, sKey)
    }

    pipe.Del(ctx, uKey)

    _, err = pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("一括削除に失敗しました: %w", err)
    }

    return nil
}