package service

import (
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"context"
	"errors"
	"time"

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

type mockSessionStore struct {
	mock.Mock
}

func (m *mockSessionStore) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	args := m.Called(ctx, session)
	return testutils.SafeGet[models.Session](args, 0), args.Error(1)
}

func(m *mockSessionStore) GetByHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	args := m.Called(ctx, tokenHash)
	return testutils.SafeGet[models.Session](args, 0), args.Error(1)
}

func(m *mockSessionStore) UpdateExpiresAt(ctx context.Context, expiresAt time.Time, id int64) error {
	args := m.Called(ctx, expiresAt, id)
	return args.Error(0)
}

func(m *mockSessionStore) DeleteByHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx,tokenHash)
	return args.Error(0)
}

type mockTweetStore struct {
	mock.Mock
}

func (m *mockTweetStore) CreateTweet(ctx context.Context, twt *models.Tweet) (*models.Tweet, error) {
	args := m.Called(ctx, twt)
	return testutils.SafeGet[models.Tweet](args, 0), args.Error(1)
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
