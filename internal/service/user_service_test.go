package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
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
        inputBody *dto.SignupRequest
        setupMock func(mu *mockUserRepository, mh *mockBcryptHasher)
        wantedErr error
        errMsg    string
    }{
        "登録成功": {
            inputBody: &dto.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {
                mh.On("Generate", "password101").Return("mock_hash", nil)
                mu.On("Create", mock.Anything, mock.MatchedBy(func(r *dto.UserRecord) bool {
                    return r.Username == "mock_user" && r.Email == "mock@example.com" && r.PasswordHash == "mock_hash"
                })).Return(&dto.UserRecord{
                    ID:           1,
                    Username:     "mock_user",
                    Email:        "mock@example.com",
                    PasswordHash: "mock_hash",
                    CreatedAt:    time.Now().UTC(),
                }, nil)
            },
            wantedErr: nil,
        },
        "必須項目が不足": {
            inputBody: &dto.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "", 
            },
            setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {},
            wantedErr: errcode.ErrRequiredFieldMissing,
        },
        "パスワードをハッシュ化に失敗": {
            inputBody: &dto.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {
                mh.On("Generate", "password101").Return("", errMockHashFailed)
            },
            wantedErr: errMockHashFailed,
            errMsg:    "パスワードをハッシュ化に失敗しました",
        },
        "データベース内部エラー": {
            inputBody: &dto.SignupRequest{
                Username: "mock_user",
                Email:    "mock@example.com",
                Password: "password101",
            },
            setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {
                mh.On("Generate", "password101").Return("mock_hash", nil)
                mu.On("Create", mock.Anything, mock.MatchedBy(func(r *dto.UserRecord) bool {
                    return r.Username == "mock_user"
                })).Return(nil, errMockInternal)
            },
            wantedErr: errMockInternal,
            errMsg:    "登録に失敗しました",
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            mu := new(mockUserRepository)
            mh := new(mockBcryptHasher)
            tt.setupMock(mu, mh)
            svc := NewUserService(mu, mh)
            ctx := context.Background()
            res, err := svc.Register(ctx, tt.inputBody.Username, tt.inputBody.Email, tt.inputBody.Password)

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

func TestExistsByEmail(t *testing.T) {
    tests := map[string]struct {
        email          string
        setupMock      func(m *mockUserRepository)
        wantErr        error
        expectedErrMsg string
    }{
        "【正常系】ユーザーが存在する場合、正常にユーザー情報を返す": {
            email: "exist@example.com",
            setupMock: func(m *mockUserRepository) {
                m.On("GetByEmail", mock.Anything, "exist@example.com").
                    Return(&dto.UserRecord{Email: "exist@example.com"}, nil)
            },
            wantErr: nil,
        },
        "【異常系】メールアドレスが空の場合、バリデーションエラーを返す": {
            email:     "",
            setupMock: func(m *mockUserRepository) {},
            wantErr:   errcode.ErrRequiredFieldMissing,
        },
        "【異常系】ユーザーが存在しない場合、認証エラーに変換して返す（ユーザー列挙防止）": {
            email: "notfound@example.com",
            setupMock: func(m *mockUserRepository) {
                m.On("GetByEmail", mock.Anything, "notfound@example.com").
                    Return(nil, errcode.ErrUserNotFound)
            },
            wantErr: errcode.ErrInvalidCredentials,
        },
        "【異常系】データベース接続エラー等の内部エラーが発生した場合": {
            email: "db@example.com",
            setupMock: func(m *mockUserRepository) {
                m.On("GetByEmail", mock.Anything, "db@example.com").
                    Return(nil, errMockInternal )
            },
            wantErr: errMockInternal,
            expectedErrMsg: "ユーザー情報の取得に失敗しました",
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            mu := new(mockUserRepository)
            tt.setupMock(mu)
            
            svc := NewUserService(mu, nil)
            
            ctx := context.Background()
            user, err := svc.existsByEmail(ctx, tt.email)

            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, user)
                if tt.expectedErrMsg != "" {
                    assert.Contains(t, err.Error(), tt.expectedErrMsg)
                }
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.email, user.Email)
            }
            mu.AssertExpectations(t)
        })
    }
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]struct {
		email     string
		password  string
		setupMock func(mu *mockUserRepository, mh *mockBcryptHasher)
		wantedErr error
		errMsg    string
	}{
		"【正常系】ログイン成功：正しい資格情報でユーザー情報を返す": {
			email:    "mock@example.com",
			password: "password123",
			setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {
				mu.On("GetByEmail", mock.Anything, "mock@example.com").Return(&dto.UserRecord{
					ID: 1, Email: "mock@example.com", PasswordHash: "hashed_pass",
				}, nil)
				mh.On("Compare", "hashed_pass", "password123").Return(nil)
			},
			wantedErr: nil,
		},
		"【異常系】必須項目不足：パスワードが空の場合": {
			email:     "mock@example.com",
			password:  "",
			setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {},
			wantedErr: errcode.ErrRequiredFieldMissing,
		},
		"【異常系】認証失敗：ユーザーが存在しない場合（ユーザー列挙防止）": {
			email:    "none@example.com",
			password: "password123",
			setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {
				mu.On("GetByEmail", mock.Anything, "none@example.com").Return(nil, errcode.ErrUserNotFound)
			},
			wantedErr: errcode.ErrInvalidCredentials,
		},
		"【異常系】認証失敗：DB接続エラー等の内部エラー": {
			email:    "test@example.com",
			password: "password123",
			setupMock: func(ms *mockUserRepository, mh *mockBcryptHasher) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "ユーザー情報の取得に失敗しました",
		},
		"【異常系】認証失敗：パスワード不一致": {
			email:    "mock@example.com",
			password: "wrong-password",
			setupMock: func(mu *mockUserRepository, mh *mockBcryptHasher) {
				mu.On("GetByEmail", mock.Anything, "mock@example.com").Return(&dto.UserRecord{
					PasswordHash: "correct_hash",
				}, nil)
				mh.On("Compare", "correct_hash", "wrong-password").Return(fmt.Errorf("hash mismatch"))
			},
			wantedErr: errcode.ErrInvalidCredentials,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mu := new(mockUserRepository)
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
    gin.SetMode(gin.TestMode)
	tests := map[string]struct {
		userID    int64
		setupMock func(mu *mockUserRepository)
		wantedErr error
		errMsg    string
	}{
		"取得成功": {
			userID: 1,
			setupMock: func(mu *mockUserRepository) {
				mu.On("GetByID", mock.Anything, int64(1)).Return(&dto.UserRecord{ID: 1}, nil)
			},
			wantedErr: nil,
		},
		"ユーザーID無効": {
			userID:    -1,
			setupMock: func(mu *mockUserRepository) {},
			wantedErr: errcode.ErrInvalidUserID,
		},
		"データベース内部エラー": {
			userID: 1,
			setupMock: func(mu *mockUserRepository) {
				mu.On("GetByID", mock.Anything, int64(1)).Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "ユーザー情報の取得に失敗しました",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mu := new(mockUserRepository)
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

func TestUpdateFollowerCount(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := []struct {
        name      string
        userID    int64
        delta     int64
		setupMock func(mu *mockUserRepository)
		wantedErr error
		errMsg    string
    }{
        {
            name: "正常系: updateに成功した",
            userID: 1,
            delta: 1,
            setupMock: func(mu *mockUserRepository) {
                mu.On("IncreaseFollower", mock.Anything, int64(1), int64(1)).Return(nil)
            }, 
        },
        {
            name: "異常系: ユーザーID無効",
            userID: 0,
            delta: 1,
            setupMock: func(mu *mockUserRepository) {},
            wantedErr: errcode.ErrInvalidUserID,
        },
        {
            name: "異常系: サーバー内部エラー",
            userID: 101,
            delta: 10,
            setupMock: func(mu *mockUserRepository) {
                mu.On("IncreaseFollower", mock.Anything, int64(101),int64(10)).Return(errMockInternal)        
            },
            wantedErr: errMockInternal,
            errMsg: "FollowerCountの更新に失敗しました",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mu := new(mockUserRepository)
            tt.setupMock(mu)

            svc := NewUserService(mu, nil)
            ctx := context.Background()
            err := svc.UpdateFollowerCount(ctx, tt.userID, tt.delta)

            if tt.wantedErr != nil {
                assert.ErrorIs(t, err, tt.wantedErr)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                } 
            } else {
                assert.NoError(t, err)
            }

            mu.AssertExpectations(t)
        })
    }
}

func TestUpdateFollowingCount(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := []struct {
        name      string
        userID    int64
        delta     int64
		setupMock func(mu *mockUserRepository)
		wantedErr error
		errMsg    string
    }{
        {
            name: "正常系: updateに成功した",
            userID: 1,
            delta: 1,
            setupMock: func(mu *mockUserRepository) {
                mu.On("IncreaseFollowing", mock.Anything, int64(1), int64(1)).Return(nil)
            }, 
        },
        {
            name: "異常系: ユーザーID無効",
            userID: 0,
            delta: 1,
            setupMock: func(mu *mockUserRepository) {},
            wantedErr: errcode.ErrInvalidUserID,
        },
        {
            name: "異常系: サーバー内部エラー",
            userID: 101,
            delta: 10,
            setupMock: func(mu *mockUserRepository) {
                mu.On("IncreaseFollowing", mock.Anything, int64(101), int64(10)).Return(errMockInternal)        
            },
            wantedErr: errMockInternal,
            errMsg: "FollowingCountの更新に失敗しました",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mu := new(mockUserRepository)
            tt.setupMock(mu)

            svc := NewUserService(mu, nil)
            ctx := context.Background()
            err := svc.UpdateFollowingCount(ctx, tt.userID, tt.delta)

            if tt.wantedErr != nil {
                assert.ErrorIs(t, err, tt.wantedErr)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                } 
            } else {
                assert.NoError(t, err)
            }

            mu.AssertExpectations(t)
        })
    }
}

func TestExist(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := []struct {
        name      string
        userID    int64
		setupMock func(mu *mockUserRepository)
		wantedErr error
        wantedRes bool
		errMsg    string
    }{
        {
            name: "正常系: 存在",
            userID: 1,
            setupMock: func(mu *mockUserRepository) {
                mu.On("Exists", mock.Anything, int64(1)).Return(true, nil)
            }, 
            wantedRes: true,
        },
        {
            name: "正常系: 存在しない",
            userID: 2,
            setupMock: func(mu *mockUserRepository) {
                mu.On("Exists", mock.Anything, int64(2)).Return(false, nil)
            },
            wantedRes: false,
        },
        {
            name: "異常系: ユーザーID無効",
            userID: 0,
            setupMock: func(mu *mockUserRepository) {},
            wantedErr: errcode.ErrInvalidUserID,
            wantedRes: false,
        },
        {
            name: "異常系: サーバー内部エラー",
            userID: 101,
            setupMock: func(mu *mockUserRepository) {
                mu.On("Exists", mock.Anything, int64(101)).Return(false, errMockInternal)        
            },
            wantedErr: errMockInternal,
            errMsg: "ユーザーの存在の確認に失敗しました",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mu := new(mockUserRepository)
            tt.setupMock(mu)

            svc := NewUserService(mu, nil)
            ctx := context.Background()
            res, err := svc.Exists(ctx, tt.userID)

            assert.Equal(t, tt.wantedRes, res)
            if tt.wantedErr != nil {
                assert.ErrorIs(t, err, tt.wantedErr)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                } 
            } else {
                assert.NoError(t, err)
            }

            mu.AssertExpectations(t)
        })
    }
}