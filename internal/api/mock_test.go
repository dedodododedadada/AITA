package api

import (
	"aita/internal/models"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

type MockSessionStore struct {
	mock.Mock
}

func safeGet[T any](args mock.Arguments, index int) *T {
	val := args.Get(index)
	if val == nil {
		return nil
	}
	return val.(*T)
}

func(m *MockUserStore) Create(ctx context.Context, req *models.SignupRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	return safeGet[models.User](args, 0), args.Error(1)
}

func(m *MockUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return safeGet[models.User](args, 0), args.Error(1)
}

func (m *MockUserStore) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return safeGet[models.User](args, 0), args.Error(1)
}	

func (m *MockSessionStore) Create(ctx context.Context, userID int64, duration time.Duration) (string, *models.Session, error) {
	args := m.Called(ctx, userID, duration)
	token, _ := args.Get(0).(string)
	return token, safeGet[models.Session](args, 1), args.Error(2)
}

func (m *MockSessionStore) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	args := m.Called(ctx, token)
	return safeGet[models.Session](args, 0), args.Error(1)
}

