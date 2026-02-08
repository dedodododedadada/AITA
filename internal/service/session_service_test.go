package service

import (
	"aita/internal/models"
	"context"
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
		setupMock func(ms *mockSessionStore, mt *mockTokenManager)
		wantedErr error
		errMsg    string
	}{
		"発行成功": {
			userID: 1,
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
				rawToken := "generated_raw_token_32_characters_long"
				hashedToken := "hashed_token"

				mt.On("Generate", 32).Return(rawToken, nil)
				mt.On("Hash", rawToken).Return(hashedToken)
				expectedSession := &models.Session{
						ID: 1,
           				UserID: 1,
            			TokenHash: hashedToken,
            			CreatedAt: time.Now().UTC(),
            			ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
				}

				ms.On("Create", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
					return s.UserID == 1 && s.TokenHash == hashedToken
				})).Return(expectedSession, nil)
			},
			wantedErr: nil,
		},
		"バリデーションエラー": {
			userID:    0,
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
			wantedErr: models.ErrRequiredFieldMissing,
		},
		"トークン生成失敗": {
			userID: 1,
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
				mt.On("Generate", 32).Return("", errMockTokenFailed)
			},
			wantedErr: errMockTokenFailed,
			errMsg:    "トークンの生成に失敗しました",
		},
		"DB保存失敗": {
			userID: 1,
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
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
			ms := new(mockSessionStore)
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
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.userID, res.Session.UserID)
				assert.Equal(t, "generated_raw_token_32_characters_long", res.Token)
				assert.Equal(t, int64(1), res.Session.ID)
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
		setupMock func(ms *mockSessionStore, mt *mockTokenManager)
		wantedErr error
		errMsg    string
	}
	tests := []testCase{
		{
			name:  "認証成功",
			token: "valid_token_that_is_long_enough_32char",
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
				tokenHash := "hashed_token"
				mt.On("Hash", mock.Anything).Return(tokenHash)
				ms.On("GetByHash", mock.Anything, tokenHash).Return(&models.Session{
					ID:        1,
					UserID:    10,
					TokenHash: "hashed_token",
					ExpiresAt: time.Now().Add(24* time.Hour).UTC(),
					CreatedAt: time.Now().UTC(),
				}, nil)
			},
			wantedErr: nil,
		},
		{
			name:      "Tokenが空",
			token:     "",
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
			wantedErr: models.ErrRequiredFieldMissing,
		},
		{
			name:      "Tokenが短すぎる",
			token:     "too_short",
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
			wantedErr: models.ErrInvalidTokenFormat,
		},
		{
			name: "Tokenが長すぎる",
			token: strings.Repeat("a", 256),
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
			wantedErr: models.ErrInvalidTokenFormat,
		},
		{
			name:  "セッションが見つからない",
			token: "unknown_token_long_enough_32char",
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
				mt.On("Hash", mock.Anything).Return("unknown_hash")
				ms.On("GetByHash", mock.Anything, "unknown_hash").Return(nil, errMockTokenFailed)
			},
			wantedErr: errMockTokenFailed,
			errMsg:    "セッションの取得に失敗しました",
		},
		{
			name:  "DB内部エラー",
			token: "valid_token_that_is_long_enough_32char",
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
				mt.On("Hash", mock.Anything).Return("hash")
				ms.On("GetByHash", mock.Anything, "hash").Return(nil, errMockTokenFailed)
			},
			wantedErr: errMockTokenFailed,
			errMsg:    "セッションの取得に失敗しました",
		},
		{
			name: "期限切れ",
			token: "valid_token_that_is_long_enough_32char",
			setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
				tokenHash := "hashed_token"
				mt.On("Hash", mock.Anything).Return(tokenHash)
				ms.On("GetByHash", mock.Anything, tokenHash).Return(&models.Session{
					ID:        1,
					UserID:    10,
					TokenHash: "hashed_token",
					ExpiresAt: time.Now().Add(-1 * time.Hour).UTC(),
					CreatedAt: time.Now().UTC(),
				}, nil)
			},
			wantedErr: models.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionStore)
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
		setupMock func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager)
		wantedErr error
		errMsg    string
	}

	validToken := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	tests := []testCase{
		{
			name:      "トークンが完全に空",
			token:     "",
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {},
			wantedErr: models.ErrRequiredFieldMissing,
		},
		{
			name:      "フォーマット不正",
			token:     "Bearer" + validToken, 
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {},
			wantedErr: models.ErrInvalidTokenFormat,
		},
		{
			name:      "プレフィックスがBearerではない",
			token:     "Basic " + validToken,
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {},
			wantedErr: models.ErrInvalidTokenFormat,
		},
		{
			name:      "Bearerヘッダーのみで内容が空",
			token:     "Bearer ",
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
			},
			wantedErr: models.ErrRequiredFieldMissing,
		},
		{
			name:      "トークン長が不足",
			token:     "Bearer short_token",
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
			},
			wantedErr: models.ErrInvalidTokenFormat,
		},
		{
			name:  "データベース内部エラー",
			token: "Bearer " + validToken,
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				ms.On("GetByHash", mock.Anything, "hashed_ok").Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg: "セッションの取得に失敗しました",
		},
		{
			name:  "【成功】標準的なBearer認証",
			token: "Bearer " + validToken,
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				session := &models.Session{
					ID: 1, 
					UserID: 10, 
					TokenHash: "hashed_ok",
					ExpiresAt: time.Now().Add(23 * time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1 * time.Hour).UTC(),
				}
				ms.On("GetByHash", mock.Anything, "hashed_ok").Return(session, nil)
				mu.On("ToMyPage", mock.Anything, int64(10)).Return(&models.User{ID: 10}, nil)
			},
			wantedErr: nil,
		},
		{
			name:  "【成功】Bearerの大小文字を区別しない",
			token: "bEaReR " + validToken,
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				session := &models.Session{
					ID: 1,
					UserID: 10, 
					TokenHash: "hashed_ok",
					ExpiresAt: time.Now().Add(23 * time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1 * time.Hour).UTC(),
				}
				ms.On("GetByHash", mock.Anything, "hashed_ok").Return(session, nil)
				mu.On("ToMyPage", mock.Anything, int64(10)).Return(&models.User{ID: 10}, nil)
			},
			wantedErr: nil,
		},
		{
			name:  "【成功】セッションの自動更新がトリガーされる",
			token: "Bearer " + validToken,
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				oldExpiry := time.Now().Add(1 * time.Hour).UTC()
				session := &models.Session{
					ID: 1, 
					UserID: 10, 
					TokenHash: "hashed_ok",
					ExpiresAt: oldExpiry,
					CreatedAt: time.Now().Add(-23 * time.Hour).UTC(),
				}
				ms.On("GetByHash", mock.Anything, "hashed_ok").Return(session, nil)
				mu.On("ToMyPage", mock.Anything, int64(10)).Return(&models.User{ID: 10}, nil)
				ms.On("UpdateExpiresAt", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
            		expected := time.Now().Add(models.SessionDuration).UTC()
           			 return t.After(oldExpiry) && t.Sub(expected).Abs() < 10*time.Second
        		}), int64(1)).Return(nil)
			},
			wantedErr: nil,
		},
		{
			name:  "【失敗】セッションは有効だがユーザーが存在しない（退会済みなど）",
			token: "Bearer " + validToken,
			setupMock: func(ms *mockSessionStore, mu *mockUserService, mt *mockTokenManager) {
				mt.On("Hash", validToken).Return("hashed_ok")
				session := &models.Session{
					ID: 1,
					UserID: 10, 
					TokenHash: "hashed_ok",
					ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
					CreatedAt: time.Now().Add(-1 * time.Hour).UTC(),
				}
				ms.On("GetByHash", mock.Anything, "hashed_ok").Return(session, nil)
				mu.On("ToMyPage", mock.Anything, int64(10)).Return(nil, models.ErrUserNotFound)
			},
			wantedErr: models.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionStore)
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
				if tt.name == "【成功】セッションの自動更新がトリガーされ、正しく更新されること" {
    			require.NoError(t, err)	
    				expectedNewExpiry := time.Now().Add(models.SessionDuration).UTC()
    				assert.WithinDuration(t, expectedNewExpiry, res.ExpiresAt, 10*time.Second)
    				assert.True(t, res.ExpiresAt.After(time.Now().Add(22*time.Hour))) 
				}
				require.NoError(t, err)
				assert.NotNil(t, res)
			}

			ms.AssertExpectations(t)
			mt.AssertExpectations(t)
			mu.AssertExpectations(t)
		})
	}
}

func TestRefreshSession(t *testing.T) {
	initialExpiry := time.Now().Add(-1 * time.Hour).UTC()
	session := &models.Session{
		ID:        1,
		UserID:    10,
		ExpiresAt: initialExpiry,
	}

	tests := []struct {
		name      string
		setupMock func(ms *mockSessionStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "セッション期限が正常に更新される",
			setupMock: func(ms *mockSessionStore) {
				ms.On("UpdateExpiresAt", mock.Anything, mock.AnythingOfType("time.Time"), int64(1)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "データベース更新エラー",
			setupMock: func(ms *mockSessionStore) {
				ms.On("UpdateExpiresAt", mock.Anything, mock.Anything, int64(1)).
					Return(errMockInternal)
			},
			wantErr: true,
			errMsg:  "セッション期限の更新に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockSessionStore)
			mt := new(mockTokenManager)
			mu := new(mockUserService)
			
			tt.setupMock(ms)
			svc := NewSessionService(ms, mu, mt)

			err := svc.refreshSession(context.Background(), session)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, session.ExpiresAt.After(initialExpiry), "ExpiresAt が更新後の方が新しくなっているべき")
				expectedExpiry := time.Now().Add(models.SessionDuration)
				assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, 10*time.Second)
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestRevoke(t *testing.T) {
    type testCase struct {
        name           string
        header         string 
        setupMock      func(ms *mockSessionStore, mt *mockTokenManager)
        wantedErr      error
        expectContains string
    }

    validToken := "valid_token_that_is_long_enough_32char"
    validHeader := "Bearer " + validToken

    tests := []testCase{
        {
            name:   "正常にログアウトできる(正規のBearerヘッダー)",
            header: validHeader,
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
                mt.On("Hash", validToken).Return("token_hash")
                ms.On("DeleteByHash", mock.Anything, "token_hash").Return(nil)
            },
        },
        {
            name:      "ヘッダーが空",
            header:    "",
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
            wantedErr: models.ErrSessionNotFound,
        },
        {
            name:      "Bearerプレフィックスがない",
            header:    validToken, 
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
            wantedErr: models.ErrSessionNotFound,
        },
        {
            name:      "Bearerの綴りがおかしい",
            header:    "Beareeee " + validToken,
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
            wantedErr: models.ErrSessionNotFound,
        },
        {
            name:   "大文字のBEARERでも認識されること",
            header: "BEARER " + validToken,
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
                mt.On("Hash", validToken).Return("token_hash")
                ms.On("DeleteByHash", mock.Anything, "token_hash").Return(nil)
            },
        },
		{
            name:      "tokenが空(Bearerのみ)",
            header:    "Bearer ",
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
            wantedErr: models.ErrSessionNotFound,
        },
		{
    		name:   "tokenが31文字(短すぎる)",
    		header: "Bearer " + strings.Repeat("a", 31),
    		setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
    		wantedErr: models.ErrSessionNotFound,
		},
		{
		    name:   "tokenが256文字(長すぎる)",
    		header: "Bearer " + strings.Repeat("a", 256),
   		    setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {},
    		wantedErr: models.ErrSessionNotFound,
		},
		{
    		name:   "tokenがちょうど32文字(境界値・成功)",
    		header: "Bearer " + strings.Repeat("a", 32),
    		setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
       			mt.On("Hash", strings.Repeat("a", 32)).Return("hash_32")
        		ms.On("DeleteByHash", mock.Anything, "hash_32").Return(nil)
    		},
    		wantedErr: nil,
		},
        {
            name:   "DBエラー時にラップされたエラーを返す",
            header: validHeader,
            setupMock: func(ms *mockSessionStore, mt *mockTokenManager) {
                mt.On("Hash", validToken).Return("token_hash")
                ms.On("DeleteByHash", mock.Anything, "token_hash").Return(errMockInternal)
            },
            wantedErr:      errMockInternal,
            expectContains: "セッションの削除に失敗しました",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ms := new(mockSessionStore)
            mt := new(mockTokenManager)
            mu := new(mockUserService)
            svc := NewSessionService(ms, mu, mt)

            tt.setupMock(ms, mt)

            err := svc.Revoke(context.Background(), tt.header)

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

