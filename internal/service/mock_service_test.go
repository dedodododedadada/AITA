package service

import (
	"aita/internal/models"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockTweetStore struct {
	mock.Mock
}

func safeGet[T any](args mock.Arguments, index int) *T {
	val := args.Get(index)
	if val == nil {
		return nil
	}
	return val.(*T)
}

func(m *MockTweetStore) CreateTweet(ctx context.Context, twt *models.Tweet) error {
	args := m.Called(ctx, twt)
	return args.Error(0)
}
