package service

import (
	"aita/internal/dto"
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"context"
	"errors"

	"github.com/stretchr/testify/mock"
)

var (
    errMockInternal = errors.New("接続拒否")
    errMockHashFailed = errors.New("暗号化内部エラー")
	errMockTokenFailed = errors.New("トークン内部エラー")
)

type mockUserStore struct {
	mock.Mock
}

func (m *mockUserStore) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}

func (m *mockUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return testutils.SafeGet[models.User](args,0), args.Error(1)
}

func (m *mockUserStore) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	return testutils.SafeGet[models.User](args,0), args.Error(1)
}

type mockSessionRepository struct {
	mock.Mock
}

func (m *mockSessionRepository) Create(ctx context.Context, sr *dto.SessionRecord) (*dto.SessionRecord, error){
	args := m.Called(ctx, sr)
	return testutils.SafeGet[dto.SessionRecord](args, 0), args.Error(1)
}

func(m *mockSessionRepository) Get(ctx context.Context, tokenHash string) (*dto.SessionRecord, error)  {
	args := m.Called(ctx, tokenHash)
	return testutils.SafeGet[dto.SessionRecord](args, 0), args.Error(1)
}

func(m *mockSessionRepository) Update(ctx context.Context, sr *dto.SessionRecord) error {
	args := m.Called(ctx, sr)
	return args.Error(0)
}

func(m *mockSessionRepository) Delete(ctx context.Context, sr *dto.SessionRecord) error{
	args := m.Called(ctx, sr)
	return args.Error(0)
}

type mockTweetStore struct {
	mock.Mock
}

func (m *mockTweetStore) CreateTweet(ctx context.Context, twt *models.Tweet) (*models.Tweet, error) {
	args := m.Called(ctx, twt)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
}

func(m *mockTweetStore) GetTweetByTweetID(ctx context.Context, tweetID int64) (*models.Tweet, error) {
	args := m.Called(ctx, tweetID)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
}

func(m *mockTweetStore) UpdateContent(ctx context.Context, newContent string,  tweetID int64) (*models.Tweet, error) {
	args := m.Called(ctx, newContent, tweetID)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
}


func(m *mockTweetStore) DeleteTweet(ctx context.Context, tweetID int64) error {
	args := m.Called(ctx, tweetID)
	return args.Error(0)
}

type mockBcryptHasher struct {
	mock.Mock
}

func (m *mockBcryptHasher) Generate(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func(m *mockBcryptHasher) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

type mockTokenManager struct {
	mock.Mock
}

func(m *mockTokenManager) Generate(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func(m *mockTokenManager) Hash(hash string) string {
	args := m.Called(hash)
	return args.String(0)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) ToMyPage(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return testutils.SafeGet[models.User](args, 0), args.Error(1)
}
