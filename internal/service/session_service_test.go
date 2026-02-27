package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIssue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]struct {
		userID    int64
		setupMock func(ms *mockSessionRepository, mt *mockTokenManager)
		wantedErr error
		errMsg    string
	}{
		"発行成功": {
			userID: 1,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				rawToken := "generated_raw_token_32_characters_long"
				hashedToken := "hashed_token"

				mt.On("Generate", 32).Return(rawToken, nil)
				mt.On("Hash", rawToken).Return(hashedToken)
				expectedRecord := &dto.SessionRecord{
					UserID:    1,
					TokenHash: hashedToken,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
				}

				ms.On("Create", mock.Anything, mock.MatchedBy(func(sr *dto.SessionRecord) bool {
					return sr.TokenHash == hashedToken
				})).Return(expectedRecord, nil)
			},
			wantedErr: nil,
		},
		"バリデーションエラー": {
			userID:    0,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {},
			wantedErr: errcode.ErrRequiredFieldMissing,
		},
		"トークン生成失敗": {
			userID: 1,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Generate", 32).Return("", errMockTokenFailed)
			},
			wantedErr: errMockTokenFailed,
			errMsg:    "トークンの生成に失敗しました",
		},
		"DB保存失敗": {
			userID: 1,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Generate", 32).Return("token", nil)
				mt.On("Hash", "token").Return("hash")
				ms.On("Create", mock.Anything, mock.Anything).Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "発行に失敗しました",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mu := new(mockUserService)
			mt := new(mockTokenManager)

			tt.setupMock(ms, mt)

			svc := NewSessionService(ms, mu, mt)
			res, err := svc.Issue(context.Background(), tt.userID)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Empty(t, res) 
			} else {
				require.NoError(t, err)
				assert.Equal(t, "generated_raw_token_32_characters_long", res.Token)
			}

			ms.AssertExpectations(t)
			mt.AssertExpectations(t)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	type testCase struct {
		name      string
		token     string
		setupMock func(ms *mockSessionRepository, mt *mockTokenManager)
		wantedErr error
		errMsg    string
	}
	tests := []testCase{
		{
			name:  "認証成功",
			token: "valid_token_that_is_long_enough_32char",
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				tokenHash := "hashed_token"
				mt.On("Hash", mock.Anything).Return(tokenHash)
				ms.On("Get", mock.Anything, tokenHash).Return(&dto.SessionRecord{
					UserID:    10,
					TokenHash: "hashed_token",
					ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
					CreatedAt: time.Now().UTC(),
				}, nil)
			},
			wantedErr: nil,
		},
		{
			name:      "Tokenが空",
			token:     "",
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {},
			wantedErr: errcode.ErrSessionNotFound,
		},
		{
			name:      "Tokenが短すぎる",
			token:     "too_short",
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {},
			wantedErr: errcode.ErrInvalidTokenFormat,
		},
		{
			name:      "Tokenが長すぎる",
			token:     strings.Repeat("a", 256),
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {},
			wantedErr: errcode.ErrInvalidTokenFormat,
		},
		{
			name:  "セッションが見つからない",
			token: "unknown_token_long_enough_32char",
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Hash", mock.Anything).Return("unknown_hash")
				ms.On("Get", mock.Anything, "unknown_hash").Return(nil, errMockTokenFailed)
			},
			wantedErr: errMockTokenFailed,
			errMsg:    "セッションの取得に失敗しました",
		},
		{
			name:  "DB内部エラー",
			token: "valid_token_that_is_long_enough_32char",
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Hash", mock.Anything).Return("hash")
				ms.On("Get", mock.Anything, "hash").Return(nil, errMockTokenFailed)
			},
			wantedErr: errMockTokenFailed,
			errMsg:    "セッションの取得に失敗しました",
		},
		{
			name:  "期限切れ",
			token: "valid_token_that_is_long_enough_32char",
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				tokenHash := "hashed_token"
				mt.On("Hash", mock.Anything).Return(tokenHash)
				ms.On("Get", mock.Anything, tokenHash).Return(&dto.SessionRecord{
					UserID:    10,
					TokenHash: "hashed_token",
					ExpiresAt: time.Now().Add(-1 * time.Hour).UTC(),
					CreatedAt: time.Now().UTC(),
				}, nil)
			},
			wantedErr: errcode.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mt := new(mockTokenManager)
			mu := new(mockUserService)

			tt.setupMock(ms, mt)

			svc := NewSessionService(ms, mu, mt)
			res, err := svc.authenticate(context.Background(), tt.token)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, res)
			} else if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, int64(10), res.UserID)
			}

			ms.AssertExpectations(t)
			mt.AssertExpectations(t)
		})
	}
}

func TestValidate(t *testing.T) {
	type testCase struct {
		name      string
		token     string
		setupMock func(ms *mockSessionRepository, mu *mockUserService, mt *mockTokenManager)
		wantedErr error
		errMsg    string
	}

	validToken := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	tests := []testCase{
		{
			name:      "【失敗】トークンが空（サービス層でのガード）",
			token:     "",
			setupMock: func(ms *mockSessionRepository, mu *mockUserService, mt *mockTokenManager) {},
			wantedErr: errcode.ErrSessionNotFound,
		},
		{
			name:  "【失敗】データベース内部エラー",
			token: validToken,
			setupMock: func(ms *mockSessionRepository, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				ms.On("Get", mock.Anything, "hashed_ok").Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "セッションの取得に失敗しました",
		},
		{
			name:  "【成功】標準的な検証",
			token: validToken,
			setupMock: func(ms *mockSessionRepository, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				record := &dto.SessionRecord{
					UserID:    10,
					TokenHash: "hashed_ok",
					ExpiresAt: time.Now().Add(23 * time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1 * time.Hour).UTC(),
				}
				ms.On("Get", mock.Anything, "hashed_ok").Return(record, nil)
				mu.On("ToMyPage", mock.Anything, int64(10)).Return(&models.User{ID: 10}, nil)
			},
			wantedErr: nil,
		},
		{
			name:  "【失敗】セッションは有効だがユーザーが存在しない（退会済みなど）",
			token: validToken,
			setupMock: func(ms *mockSessionRepository, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				record := &dto.SessionRecord{
					UserID:    10,
					TokenHash: "hashed_ok",
					ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1 * time.Hour).UTC(),
				}
				ms.On("Get", mock.Anything, "hashed_ok").Return(record, nil)
				mu.On("ToMyPage", mock.Anything, int64(10)).Return(nil, errcode.ErrUserNotFound)
			},
			wantedErr: errcode.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mt := new(mockTokenManager)
			mu := new(mockUserService)

			tt.setupMock(ms, mu, mt)
			svc := NewSessionService(ms, mu, mt)

			res, err := svc.Validate(context.Background(), tt.token)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, res)

				if tt.name == "【成功】セッションの自動更新がトリガーされる" {
					expectedNewExpiry := time.Now().Add(SessionDuration).UTC()
					assert.WithinDuration(t, expectedNewExpiry, res.ExpiresAt, 10*time.Second)
				}
			}

			ms.AssertExpectations(t)
			mt.AssertExpectations(t)
			mu.AssertExpectations(t)
		})
	}
}
func TestExpirationCheck(t *testing.T) {
	tests := []struct {
		name		 	string
		setupExpiresAt  time.Time
		setupCreatedAt  time.Time
		expectedErr  	error
	} {
		{
			name: "正常系：期限内",
			setupExpiresAt: time.Now().Add(24*time.Hour).UTC(),
        	setupCreatedAt: time.Now().Add(-48*time.Hour).UTC(),
			expectedErr: nil,
		},
		{
			name: "異常系：なし",
			expectedErr: errcode.ErrRequiredFieldMissing,
		},
		{
			name: "異常系：期限切れ",
        	setupExpiresAt: time.Now().Add(-1*time.Hour).UTC(),
        	setupCreatedAt: time.Now().Add(-96*time.Hour).UTC(),
			expectedErr: errcode.ErrSessionExpired,
		},
		{
			name: "異常系：最長時間超え",
        	setupExpiresAt: time.Now().Add(1*time.Hour).UTC(),
        	setupCreatedAt: time.Now().Add(-8*24*time.Hour).UTC(),
			expectedErr: errcode.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mt := new(mockTokenManager)
			mu := new(mockUserService)
			svc := NewSessionService(ms, mu, mt)
			err := svc.expirationCheck(tt.setupExpiresAt, tt.setupCreatedAt)
			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.expectedErr)
			}
		})
	}
}

func TestShouldRefresh(t *testing.T) {
	tests := []struct {
		name			string
		setupExpiresAt  time.Time
		setupCreatedAt  time.Time
		expectedBool 	bool
		expectedErr 	error
	}{
		{
			name: "正常系：期限延長される必要なし",
			setupExpiresAt: time.Now().Add(24*time.Hour).UTC(),
        	setupCreatedAt: time.Now().Add(-48*time.Hour).UTC(),
			expectedBool: false,
			expectedErr: nil,
		},
		{
			name: "正常系：期限延長される必要がある",
        	setupExpiresAt: time.Now().Add(12*time.Hour).UTC(),
        	setupCreatedAt: time.Now().Add(-24*6*time.Hour).UTC(),
			expectedBool: true,
			expectedErr: nil,
		},
		{
			name: "異常系：パラメーター異常",
        	setupExpiresAt: time.Now().Add(12*time.Hour).UTC(),
        	setupCreatedAt: time.Now().Add(12*6*time.Hour).UTC(),
			expectedBool: false,
			expectedErr:  errcode.ErrSessionExpired,
		},
		{
			name: "異常系：なし",
			expectedBool: false,
			expectedErr:  errcode.ErrRequiredFieldMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mt := new(mockTokenManager)
			mu := new(mockUserService)

			svc := NewSessionService(ms, mu, mt)

			foundBool, err := svc.ShouldRefresh(tt.setupExpiresAt, tt.setupCreatedAt)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.Equal(t, tt.expectedBool, foundBool)
			}

		})
	}
}

func TestExecuteRefresh(t *testing.T) {
	now := time.Now().UTC()
	initialExpiry := now.Add(1 * time.Hour).UTC()
	userID := 10
	token := "sbjbabfhbanjkaflansflasbfabfb223ebfhb"
	tokenHash := "valid_token_hash"
	detailErr := fmt.Errorf("セッション取得後に返却されたオブジェクトが nil です")
	sr := &dto.SessionRecord{
		UserID:    int64(userID),
		TokenHash: tokenHash,
		ExpiresAt: initialExpiry,
		CreatedAt: now.Add(-72*time.Hour).UTC(),
	}

	srMax := &dto.SessionRecord{
        UserID:    10,
        TokenHash: tokenHash,
        ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
        CreatedAt: time.Now().Add(-MaxSessionLife).Add(1 * time.Hour).UTC(), 
    }
	tests := []struct {
		name      	string
		setupuserID int64
		setupToken 	string
		setupMock 	func(ms *mockSessionRepository,mt *mockTokenManager)
		expectErr  	error
		check       func(t *testing.T, sr *dto.SessionRecord)
		errMsg    	string
	} {
		{
			name: "正常系：セッション期限が正常に更新される(newExpiry)",
			setupuserID: 10,
			setupToken: token,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Hash", token).Return(tokenHash, nil)
				ms.On("Get", mock.Anything, tokenHash).Return(sr, nil)
				ms.On("Update", mock.Anything,mock.Anything).Return(nil)
			},
			check: func(t *testing.T, sr *dto.SessionRecord) {
           		expected := time.Now().Add(SessionDuration).UTC()
            	assert.WithinDuration(t, expected, sr.ExpiresAt, 2*time.Second)
       		 },
		},
		{
			name: "正常系：セッション期限が正常に更新される(maxExpiry)",
			setupuserID: 10,
			setupToken: token,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Hash", token).Return(tokenHash, nil)
				ms.On("Get", mock.Anything, tokenHash).Return(srMax, nil)
				ms.On("Update", mock.Anything,mock.Anything).Return(nil)
			},
			check: func(t *testing.T, sr *dto.SessionRecord) {
            	expectedMax := sr.CreatedAt.Add(MaxSessionLife).UTC()
            	assert.True(t, sr.ExpiresAt.Equal(expectedMax))
        	},
		},
		{
			name: "異常系：recordなし",
			setupuserID: 100,
			setupToken: token,
			setupMock: func(ms *mockSessionRepository, mt *mockTokenManager) {
				mt.On("Hash", token).Return(tokenHash, nil)
				ms.On("Get", mock.Anything, tokenHash).Return(nil, detailErr)
			},
			expectErr: detailErr,
			errMsg: "セッションの取得に失敗しました",
		},
		{
			name: "異常系：データベース更新エラー",
			setupuserID: 10,
			setupToken: token,
			setupMock: func(ms *mockSessionRepository,  mt *mockTokenManager) {
				mt.On("Hash", token).Return(tokenHash, nil)
				ms.On("Get", mock.Anything, tokenHash).Return(sr, nil)
				ms.On("Update", mock.Anything, mock.Anything).Return(errMockInternal)
			},
			expectErr: errMockInternal,
			errMsg:  "セッション期限の更新に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mt := new(mockTokenManager)
			mu := new(mockUserService)

			tt.setupMock(ms, mt)
			svc := NewSessionService(ms, mu, mt)

			err := svc.executeRefresh(context.Background(), tt.setupToken)
			

			if tt.expectErr != nil {
				if tt.errMsg != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errMsg)
				} else {
					assert.ErrorIs(t, err, tt.expectErr)
				}
			} else {
				call := ms.Calls[len(ms.Calls)-1] 
    			updatedSR := call.Arguments.Get(1).(*dto.SessionRecord)
    			if tt.check != nil {
       				 tt.check(t, updatedSR)
    			}
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestRefreshAsyncErrorHandling(t *testing.T) {
    ms := new(mockSessionRepository)
    mt := new(mockTokenManager)
    mu := new(mockUserService)
    
    svc := NewSessionService(ms, mu, mt)

    testToken := "valid-token-at-least-32-characters-long"
    testHash := "mocked-hash"

	mt.On("Hash", testToken).Return(testHash, nil)
    ms.On("Get", mock.Anything, testHash).Return(nil, errMockInternal)

    svc.RefreshAsync(testToken)

    time.Sleep(50 * time.Millisecond)
    ms.AssertExpectations(t)
    mt.AssertExpectations(t)
    
    t.Log("Async task failed as expected, but service is still alive.")
}

func TestRefreshAsyncSuccess(t *testing.T) {
    ms := new(mockSessionRepository)
    mt := new(mockTokenManager)
    mu := new(mockUserService)
    svc := NewSessionService(ms, mu, mt)

    testToken := "valid-token-at-least-32-characters-long"
    testHash := "mocked-hash"
    userID := int64(123)

    mt.On("Hash", testToken).Return(testHash, nil)
    ms.On("Get", mock.Anything, testHash).Return(&dto.SessionRecord{
        UserID: userID,
        TokenHash: testHash,
        ExpiresAt: time.Now().Add(1 * time.Hour),
        CreatedAt: time.Now().Add(-1 * time.Hour),
    }, nil)

    ms.On("Update", mock.Anything, mock.MatchedBy(func(r *dto.SessionRecord) bool {
        return r.UserID == userID && r.ExpiresAt.After(time.Now())
    })).Return(nil)

    svc.RefreshAsync(testToken)

    time.Sleep(50 * time.Millisecond)


    ms.AssertExpectations(t)
    mt.AssertExpectations(t)
}

func TestRefreshAsync_Isolation(t *testing.T) {
    ms := new(mockSessionRepository)
    mt := new(mockTokenManager)
    mu := new(mockUserService)
    svc := NewSessionService(ms, mu, mt)

    mt.On("Hash", mock.Anything).Return("hash", nil)
    ms.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
        time.Sleep(200 * time.Millisecond) 
    }).Return(&dto.SessionRecord{UserID: 123}, nil)
    ms.On("Update", mock.Anything, mock.MatchedBy(func(r *dto.SessionRecord) bool {
        return r.UserID == 123 && r.ExpiresAt.After(time.Now())})).Return(nil)

    startTime := time.Now()

    svc.RefreshAsync("token-at-least-32-characters-long-xxx")

    duration := time.Since(startTime)
    assert.Less(t, duration, 10*time.Millisecond, "メインプロセスがブロックされています。非同期の隔離に失敗しました。")

    t.Logf("メインプロセスの実行時間: %v - 即時レスポンスを確認しました。", duration)

    time.Sleep(500 * time.Millisecond)
    ms.AssertExpectations(t)
}

func TestRevoke(t *testing.T) {
	type testCase struct {
		name           string
		userID         int64
		token          string
		setupMock      func(ms *mockSessionRepository,  mt *mockTokenManager)
		wantedErr      error
		expectContains string
	}

	validToken := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	tests := []testCase{
		{
			name:  "【成功】正常にログアウトできる",
			userID: int64(101),
			token: validToken,
			setupMock: func(ms *mockSessionRepository,  mt *mockTokenManager) {
				check := &dto.SessionRecord{
					UserID: 101,
					TokenHash: "token_hash",
					ExpiresAt: time.Now().Add(24*time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1*time.Hour).UTC(),
				}
				mt.On("Hash", validToken).Return("token_hash", nil)
				ms.On("Get", mock.Anything, "token_hash").Return(check, nil)
				ms.On("Delete", mock.Anything, mock.MatchedBy(func(sr *dto.SessionRecord) bool {
					return sr.UserID == check.UserID && sr.TokenHash == check.TokenHash
				})).Return(nil)
			},
			wantedErr: nil,
		},
		{
			name:      "【失敗】無効なユーザーID",
			userID : int64(101),
			token: validToken,
			setupMock: func(ms *mockSessionRepository,  mt *mockTokenManager) {
				illCheck := &dto.SessionRecord{
					UserID: 200,
					TokenHash: "ill_token_hash",
					ExpiresAt: time.Now().Add(24*time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1*time.Hour).UTC(),
				}
				mt.On("Hash", validToken).Return("ill_token_hash", nil)
				ms.On("Get", mock.Anything, "ill_token_hash").Return(illCheck, nil)
			},
			wantedErr: errcode.ErrForbidden,
		},
		{
			name:  "【失敗】DBエラー時にラップされたエラーを返す",
			userID : int64(101),
			token: validToken,
			setupMock: func(ms *mockSessionRepository,  mt *mockTokenManager){
				check := &dto.SessionRecord{
					UserID: 101,
					TokenHash: "token_hash",
					ExpiresAt: time.Now().Add(24*time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1*time.Hour).UTC(),
				}
				mt.On("Hash", validToken).Return("token_hash", nil)
				ms.On("Get", mock.Anything, "token_hash").Return(check, nil)
				ms.On("Delete", mock.Anything, mock.Anything).Return(errMockInternal)
			},
			wantedErr:      errMockInternal,
			expectContains: "セッションの削除に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionRepository)
			mt := new(mockTokenManager)
			mu := new(mockUserService)
			svc := NewSessionService(ms, mu, mt)

			tt.setupMock(ms, mt)

			err := svc.Revoke(context.Background(), tt.userID, tt.token)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				if tt.expectContains != "" {
					assert.Contains(t, err.Error(), tt.expectContains)
				}
			} else {
				assert.NoError(t, err)
			}
			ms.AssertExpectations(t)
			mt.AssertExpectations(t)
		})
	}
}
