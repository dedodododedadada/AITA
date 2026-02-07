package api

import (
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"context"

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

func(m *mockUserService) Register(ctx context.Context, req *models.SignupRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}

func(m *mockUserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}

func (m *mockUserService) ToMyPage(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}	

func (m *mockSessionService) Issue(ctx context.Context, userID int64) (*models.IssueResponse, error)  {
	args := m.Called(ctx, userID)
	return testutils.SafeGet[models.IssueResponse](args, 0), args.Error(1)
}

func (m *mockSessionService)Validate(ctx context.Context, token string) (*models.Session, error) {
	args := m.Called(ctx, token)
	return testutils.SafeGet[models.Session](args, 0), args.Error(1)
}

func(m *mockTweetService) PostTweet(ctx context.Context, userID int64, req *models.CreateTweetRequest) (*models.Tweet, error) {
	args := m.Called(ctx, userID, req)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
}
