package api

import (
	"aita/internal/dto"
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

type mockSessionService struct {
	mock.Mock
}

type mockTweetService struct {
	mock.Mock
}

func (m *mockUserService) Register(ctx context.Context, username string, email string, password string) (*models.User, error)  {
	args := m.Called(ctx, username, email, password)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}

func (m *mockUserService) ToMyPage(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}

func (m *mockSessionService)Issue(ctx context.Context, userID int64) (*dto.SessionResponse, error) {
	args := m.Called(ctx, userID)
	return testutils.SafeGet[dto.SessionResponse](args, 0), args.Error(1)
}

func (m *mockSessionService)Revoke(ctx context.Context, userID int64, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *mockSessionService) Validate(ctx context.Context, token string) (*dto.SessionResponse, error) {
	args := m.Called(ctx, token)
	return testutils.SafeGet[dto.SessionResponse](args, 0), args.Error(1)
}

func (m *mockSessionService) ShouldRefresh(expiresAt, createdAt time.Time) (bool, error) {
	args := m.Called(expiresAt, createdAt)
	return args.Bool(0), args.Error(1)
}

func (m *mockSessionService) RefreshAsync(token string) {
	_ = m.Called(token)
}

func (m *mockTweetService) PostTweet(ctx context.Context, userID int64, content string, imageURL *string) (*models.Tweet, error) {
	args := m.Called(ctx, userID, content, imageURL)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
}

func (m *mockTweetService) FetchTweet(ctx context.Context, tweetID int64) (*models.Tweet, error) {
	args := m.Called(ctx, tweetID)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
}


func (m *mockTweetService) EditTweet(ctx context.Context, newContent string, tweetID int64, userID int64)  (*models.Tweet, bool, error) {
	args := m.Called(ctx,newContent, tweetID, userID)
	return testutils.SafeGet[models.Tweet](args, 0), args.Bool(1), args.Error(2)
}

func (m *mockTweetService) RemoveTweet(ctx context.Context, tweetID int64, userID int64) error {
	args := m.Called(ctx, tweetID, userID)
	return args.Error(0)
}