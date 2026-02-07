package service

import (
	"aita/internal/models"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := map[string]struct {
        inputBody *models.SignupRequest
        setupMock func(mu *mockUserStore, mh *mockBcryptHasher)
        wantedErr error
        errMsg    string 
    }{
        "登録成功": {
            inputBody: &models.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {
                mh.On("Generate", "password101").Return("mock_hash", nil)
                mu.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
                    return user.Username == "mock_user" && user.Email == "mock@example.com" && user.PasswordHash == "mock_hash"
                })).Return(&models.User{
                    ID: 1, 
                    Username: "mock_user", 
                    Email: "mock@example.com", 
                    PasswordHash: "mock_hash", 
                    CreatedAt: time.Now().UTC(),
                }, nil)},
            wantedErr: nil,
        },
        "必須項目が不足": {
            inputBody: &models.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrRequiredFieldMissing,
        },
        "ユーザー名形式が正しくない": {
            inputBody: &models.SignupRequest{
                Username: "m",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrInvalidUsernameFormat,
        },
        "パスワード形式が正しくない": {
            inputBody: &models.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "123",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrInvalidPasswordFormat,
        },
        "メールアドレス形式が正しくない": {
            inputBody: &models.SignupRequest{
                Username: "mock_user",
                Email:    "mockxamplecom",
                Password: "password101",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrInvalidEmailFormat,
        },
        "パスワードをハッシュ化に失敗": {
            inputBody: &models.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {
                mh.On("Generate", "password101").Return("", errMockHashFailed)
            },
            wantedErr: errMockHashFailed,
            errMsg: "パスワードをハッシュ化に失敗しました",
        },
        "データベース内部エラー": {
            inputBody: &models.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {
                mh.On("Generate", "password101").Return("mock_hash", nil)
                mu.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
                    return user.Username == "mock_user"
                })).Return(nil, errMockInternal)
            },
            wantedErr: errMockInternal,
            errMsg: "登録に失敗しました",
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            mu := new(mockUserStore)
            mh := new(mockBcryptHasher)
            tt.setupMock(mu, mh)
            svc := NewUserService(mu, mh)
            ctx := context.Background()
            res, err := svc.Register(ctx, tt.inputBody)

            if tt.wantedErr != nil {
                assert.ErrorIs(t, err, tt.wantedErr)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
                assert.Nil(t, res)
            } else {
                now := time.Now().UTC()
                require.NoError(t, err)
                require.NotNil(t, res)
                assert.Equal(t, tt.inputBody.Username, res.Username)
                assert.Equal(t, tt.inputBody.Email, res.Email)
                assert.NotZero(t, res.ID)
                assert.Equal(t, int64(1), res.ID)
                assert.False(t, res.CreatedAt.IsZero())
                assert.WithinDuration(t, now, res.CreatedAt, 10*time.Second)
                assert.Equal(t, time.UTC, res.CreatedAt.Location())
            }

            mh.AssertExpectations(t)
            mu.AssertExpectations(t)
        })
    }
}

func TestLogin(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := map[string]struct {
        email     string
        password  string
        setupMock func(mu *mockUserStore, mh *mockBcryptHasher)
        wantedErr error
        errMsg    string
    }{
        "ログイン成功": {
            email:    "mock@example.com",
            password: "password123",
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {
                mu.On("GetByEmail", mock.Anything, "mock@example.com").Return(&models.User{
                    ID: 1, Email: "mock@example.com", PasswordHash: "hashed_pass",
                }, nil)
                mh.On("Compare", "hashed_pass", "password123").Return(nil)
            },
            wantedErr: nil,
        },
        "必須項目が不足": {
            email:     "mock@example.com",
            password:  "",
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrRequiredFieldMissing,
        },
        "メールアドレス形式が正しくない": {
            email:     "invalid-email",
            password:  "password123",
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrInvalidEmailFormat,
        },
        "パスワード形式が正しくない": {
            email:     "mock@example.com",
            password:  "12", 
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {},
            wantedErr: models.ErrInvalidPasswordFormat,
        },
        "認証失敗(ユーザー不在)": {
            email:    "none@example.com",
            password: "password123",
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {
                mu.On("GetByEmail", mock.Anything, "none@example.com").Return(nil, models.ErrUserNotFound)
            },
            wantedErr: models.ErrInvalidCredentials,
        },
        "認証失敗(DB接続失敗など)": {
            email:    "test@example.com",
            password: "password123",
            setupMock: func(ms *mockUserStore, mh *mockBcryptHasher) {
                ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errMockInternal)
            },
            wantedErr: errMockInternal,
            errMsg: "ユーザー情報の取得に失敗しました",
        },
           "認証失敗(パスワード不一致)": {
            email:    "mock@example.com",
            password: "wrong-password",
            setupMock: func(mu *mockUserStore, mh *mockBcryptHasher) {
                mu.On("GetByEmail", mock.Anything, "mock@example.com").Return(&models.User{
                    PasswordHash: "correct_hash",
                }, nil)
                mh.On("Compare", "correct_hash", "wrong-password").Return(fmt.Errorf("hash mismatch"))
            },
            wantedErr: models.ErrInvalidCredentials,
        },
    }
     

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            mu := new(mockUserStore)
            mh := new(mockBcryptHasher)
            tt.setupMock(mu, mh)
            
            svc := NewUserService(mu, mh)
            ctx := context.Background()
            res, err := svc.Login(ctx, tt.email, tt.password)

            if tt.wantedErr != nil {
                assert.ErrorIs(t, err, tt.wantedErr)
                if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
                assert.Nil(t, res)
            } else {
                require.NoError(t, err)
                require.NotNil(t, res)
                assert.Equal(t, tt.email, res.Email)
            }
            mu.AssertExpectations(t)
            mh.AssertExpectations(t)
        })
    }
}

func TestToMyPage(t *testing.T) {
    tests := map[string]struct {
        userID    int64
        setupMock func(mu *mockUserStore)
        wantedErr error
        errMsg    string 
    }{
        "取得成功": {
            userID: 1,
            setupMock: func(mu *mockUserStore) {
                mu.On("GetByID", mock.Anything, int64(1)).Return(&models.User{ID: 1}, nil)
            },
            wantedErr: nil,
        },
        "ユーザーID無効": {
            userID: -1,
            setupMock: func(mu *mockUserStore) {},
            wantedErr: models.ErrInvalidUserID,
        },
        "データベース内部エラー": {
            userID: 1,
            setupMock: func(mu *mockUserStore) {
                mu.On("GetByID", mock.Anything, int64(1)).Return(nil, errMockInternal)
            },
            wantedErr: errMockInternal,
            errMsg: "ユーザー情報の取得に失敗しました",
        },
    }
    
    for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mu := new(mockUserStore)
			mh := new(mockBcryptHasher)
			tt.setupMock(mu)

			svc := NewUserService(mu, mh)
			res, err := svc.ToMyPage(context.Background(), tt.userID)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.userID, res.ID)
			}

			mu.AssertExpectations(t)
		})
	}
}
