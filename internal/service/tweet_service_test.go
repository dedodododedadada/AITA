package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostTweet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	imageURL := utils.StringPtr("https://example.com/mock.jpg")
	tests := []struct {
		name      string
		userID    int64
		inputBody *dto.CreateTweetRequest
		setupMock func(mt *mockTweetStore)
		wantedErr error
		errMsg    string
	}{
		{
			name:   "【正常系】ツイート投稿成功",
			userID: 101,
			inputBody: &dto.CreateTweetRequest{
				Content:  "Hello world",
				ImageURL: imageURL,
			},
			setupMock: func(mt *mockTweetStore) {
				expectedTweet := &models.Tweet{
					ID:        1,
					UserID:    101,
					Content:   "Hello world",
					ImageURL:  imageURL,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				mt.On("CreateTweet", mock.Anything, mock.MatchedBy(func(t *models.Tweet) bool {
					return t.UserID == 101 && t.Content == "Hello world" && t.ImageURL == imageURL
				})).Return(expectedTweet, nil)
			},
			wantedErr: nil,
		},
		{
			name:   "【異常系】無効なユーザーID",
			userID: -1,
			inputBody: &dto.CreateTweetRequest{
				Content: "Hello world",
			},
			setupMock: func(mt *mockTweetStore) {},
			wantedErr: errcode.ErrInvalidUserID,
		},
		{
			name:   "【異常系】コンテンツが空（必須項目不足）",
			userID: 101,
			inputBody: &dto.CreateTweetRequest{
				Content: "",
			},
			setupMock: func(mt *mockTweetStore) {},
			wantedErr: errcode.ErrRequiredFieldMissing,
		},
		{
			name:   "【異常系】データベースエラー（挿入失敗）",
			userID: 99999,
			inputBody: &dto.CreateTweetRequest{
				Content:  "Hello world",
				ImageURL: imageURL,
			},
			setupMock: func(mt *mockTweetStore) {
				mt.On("CreateTweet", mock.Anything, mock.MatchedBy(func(t *models.Tweet) bool {
					return t.UserID == 99999 && t.Content == "Hello world"
				})).Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "ツイートの挿入に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetStore)
			tt.setupMock(mt)
			svc := NewTweetService(mt)
			ctx := context.Background()

			res, err := svc.PostTweet(ctx, tt.userID, tt.inputBody.Content, tt.inputBody.ImageURL)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr, "期待されるエラータイプが一致します")
				assert.Nil(t, res, "エラーが発生した場合は、レスポンスはnilであるべきです")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.inputBody.Content, res.Content)
				assert.Equal(t, tt.inputBody.ImageURL, res.ImageURL)
				assert.Equal(t, tt.userID, res.UserID)
				assert.Equal(t, time.UTC, res.CreatedAt.Location())
				assert.Equal(t, time.UTC, res.UpdatedAt.Location())
			}

			mt.AssertExpectations(t)
		})
	}
}

func TestFetchTweet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		inputTweetID int64
		setupMock    func(mt *mockTweetStore)
		wantedErr    error
		errMsg       string
	}{
		{
			name:         "正常系: ツイートの取得に成功した",
			inputTweetID: 101,
			setupMock: func(mt *mockTweetStore) {
				expectedTweet := &models.Tweet{
					ID:        101,
					Content:   "mock",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(expectedTweet, nil)
			},
			wantedErr: nil,
		},
		{
			name:         "異常系:パラメーターエラー, 無効なツイートID",
			inputTweetID: 0,
			setupMock:    func(mt *mockTweetStore) {},
			wantedErr:    errcode.ErrInvalidTweetID,
		},
		{
			name:         "異常系：データベースエラー,ツイートが存在しない場合",
			inputTweetID: 101,
			setupMock: func(mt *mockTweetStore) {
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(nil, errcode.ErrTweetNotFound)
			},
			wantedErr: errcode.ErrTweetNotFound,
			errMsg:    "ツイート情報の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetStore)
			tt.setupMock(mt)
			svc := NewTweetService(mt)
			ctx := context.Background()
			res, err := svc.FetchTweet(ctx, tt.inputTweetID)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				assert.Nil(t, res)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), "ツイート情報の取得に失敗しました")
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.inputTweetID, res.ID)
				assert.Equal(t, "mock", res.Content)
				assert.Equal(t, time.UTC, res.CreatedAt.Location())
				assert.Equal(t, time.UTC, res.UpdatedAt.Location())
			}

			mt.AssertExpectations(t)
		})
	}
}

func TestToMyTweet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tweetID := int64(101)
	userID := int64(102)
	tests := []struct {
		name         string
		inputTweetID int64
		inputUserID  int64
		setupMock    func(mt *mockTweetStore)
		wantedErr    error
		errMsg       string
	}{
		{
			name:         "正常系: 自分のツイートの取得に成功した",
			inputTweetID: 101,
			inputUserID:  102,
			setupMock: func(mt *mockTweetStore) {
				expectedTweet := &models.Tweet{
					ID:        tweetID,
					UserID:    userID,
					Content:   "mock",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				mt.On("GetTweetByTweetID", mock.Anything, tweetID).Return(expectedTweet, nil)
			},
			wantedErr: nil,
		},
		{
			name:         "異常系:パラメーターエラー, 無効なツイートID",
			inputTweetID: 0,
			inputUserID:  userID,
			setupMock:    func(mt *mockTweetStore) {},
			wantedErr:    errcode.ErrInvalidTweetID,
		},
		{
			name:         "異常系:パラメーターエラー, 無効なuserID",
			inputTweetID: tweetID,
			inputUserID:  0,
			setupMock:    func(mt *mockTweetStore) {},
			wantedErr:    errcode.ErrInvalidUserID,
		},
		{
			name:         "異常系：データベースエラー,ツイートが存在しない場合",
			inputTweetID: tweetID,
			inputUserID:  userID,
			setupMock: func(mt *mockTweetStore) {
				mt.On("GetTweetByTweetID", mock.Anything, tweetID).Return(nil, errcode.ErrUserNotFound)
			},
			wantedErr: errcode.ErrUserNotFound,
			errMsg:    "ツイート情報の取得に失敗しました",
		},
		{
			name:         "異常系：権限がない",
			inputTweetID: 101,
			inputUserID:  102,
			setupMock: func(mt *mockTweetStore) {
				abnormalTweet := &models.Tweet{
					ID:        tweetID,
					UserID:    999,
					Content:   "mock",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				mt.On("GetTweetByTweetID", mock.Anything, tweetID).Return(abnormalTweet, nil)
			},
			wantedErr: errcode.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetStore)
			tt.setupMock(mt)
			svc := NewTweetService(mt)
			ctx := context.Background()
			res, err := svc.ToMyTweet(ctx, tt.inputTweetID, tt.inputUserID)

			if tt.wantedErr != nil {
				assert.ErrorIs(t, err, tt.wantedErr)
				assert.Nil(t, res)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), "ツイート情報の取得に失敗しました")
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.inputTweetID, res.ID)
				assert.Equal(t, "mock", res.Content)
				assert.Equal(t, time.UTC, res.CreatedAt.Location())
				assert.Equal(t, time.UTC, res.UpdatedAt.Location())
			}

			mt.AssertExpectations(t)
		})
	}
}

func TestEditTweet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		inputContent string
		inputTweetID int64
		inputUserID  int64
		setupMock    func(mt *mockTweetStore)
		wantedErr    error
		errMsg       string
	}{
		{
			name:         "正常系：10分以内の編集が正常に完了する",
			inputContent: "updated content",
			inputTweetID: 101,
			inputUserID:  102,
			setupMock: func(mt *mockTweetStore) {
				existingTweet := &models.Tweet{
					ID:        101,
					UserID:    102,
					Content:   "content",
					CreatedAt: time.Now().UTC().Add(-5 * time.Minute),
					UpdatedAt: time.Now().UTC().Add(-5 * time.Minute),
				}
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(existingTweet, nil)

				updatedTweet := &models.Tweet{
					ID: 101, 
					UserID: 102, 
					Content: "updated content",	
					CreatedAt: time.Now().UTC().Add(-5 * time.Minute),
					UpdatedAt: time.Now().UTC(),
					IsEdited: true,
				}
				mt.On("UpdateContent", mock.Anything, "updated content", int64(101)).Return(updatedTweet, nil)
			},
			wantedErr: nil,
		},
		{
			name:         "異常系：投稿から10分以上経過しているため編集不可",
			inputContent: "too late",
			inputTweetID: 101,
			inputUserID:  102,
			setupMock: func(mt *mockTweetStore) {
				oldTweet := &models.Tweet{
					ID:        101,
					UserID:    102,
					Content:   "content",
					CreatedAt: time.Now().UTC().Add(-11 * time.Minute),
				}
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(oldTweet, nil)
			},
			wantedErr: errcode.ErrEditTimeExpired,
		},
		{
			name:         "異常系：他のユーザーのツイートは編集不可（権限エラー）",
			inputContent: "hack",
			inputTweetID: 101,
			inputUserID:  999,
			setupMock: func(mt *mockTweetStore) {
				existingTweet := &models.Tweet{
					ID:        101,
					UserID:    102,
					Content:   "content",
					CreatedAt: time.Now().UTC(),
				}
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(existingTweet, nil)
			},
			wantedErr: errcode.ErrForbidden,
		},
		{
			name:         "正常系：内容に変更がない場合は早期リターンする",
			inputContent: "same content",
			inputTweetID: 101,
			inputUserID:  102,
			setupMock: func(mt *mockTweetStore) {
				existing := &models.Tweet{
					ID:        101,
					UserID:    102,
					Content:   "same content",
					CreatedAt: time.Now().UTC(),
				}
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(existing, nil)
			},
			wantedErr: nil,
		},
		{
			name:         "異常系：データベースエラーによる更新失敗",
			inputContent: "new content",
			inputTweetID: 101,
			inputUserID:  102,
			setupMock: func(mt *mockTweetStore) {
				existingTweet := &models.Tweet{
					ID:        101,
					UserID:    102,
					Content:   "content",
					CreatedAt: time.Now().UTC(),
				}
				mt.On("GetTweetByTweetID", mock.Anything, int64(101)).Return(existingTweet, nil)
				mt.On("UpdateContent", mock.Anything, "new content", int64(101)).Return(nil, errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "ツイート編集に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetStore)
			tt.setupMock(mt)
			svc := NewTweetService(mt)

			res, bool, err := svc.EditTweet(context.Background(), tt.inputContent, tt.inputTweetID, tt.inputUserID)

			if tt.wantedErr != nil {
				if tt.errMsg != "" {
					assert.ErrorIs(t, err, tt.wantedErr)
					assert.Contains(t, err.Error(), tt.errMsg)
					t.Logf("エラーは %v\n", err)
					assert.Equal(t, false, bool)
				}
				assert.ErrorIs(t, err, tt.wantedErr)
				assert.Nil(t, res)
				assert.Equal(t, false, bool)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, tt.inputContent, res.Content)
			}

			mt.AssertExpectations(t)
		})
	}
}

func TestRemoveTweet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		inputTweetID int64
		inputUserID  int64
		setupMock    func(mt *mockTweetStore)
		wantedErr    error
		errMsg       string
	}{
		{
			name:         "正常系: 自分のツイートが正常に削除される",
			inputTweetID: 201,
			inputUserID:  202,
			setupMock: func(mt *mockTweetStore) {
				existingTweet := &models.Tweet{ID: 201, UserID: 202}
				mt.On("GetTweetByTweetID", mock.Anything, int64(201)).Return(existingTweet, nil)
				mt.On("DeleteTweet", mock.Anything, int64(201)).Return(nil)
			},
			wantedErr: nil,
		},
		{
			name:         "異常系: ツイートが存在しない場合は削除不可",
			inputTweetID: 404,
			inputUserID:  202,
			setupMock: func(mt *mockTweetStore) {
				mt.On("GetTweetByTweetID", mock.Anything, int64(404)).Return(nil, errcode.ErrTweetNotFound)
			},
			wantedErr: errcode.ErrTweetNotFound,
		},
		{
			name:         "異常系: 他人のツイートは削除不可（権限エラー）",
			inputTweetID: 201,
			inputUserID:  999,
			setupMock: func(mt *mockTweetStore) {
				existingTweet := &models.Tweet{ID: 201, UserID: 202}
				mt.On("GetTweetByTweetID", mock.Anything, int64(201)).Return(existingTweet, nil)
			},
			wantedErr: errcode.ErrForbidden,
		},
		{
			name:         "異常系: DBエラーによる削除失敗",
			inputTweetID: 201,
			inputUserID:  202,
			setupMock: func(mt *mockTweetStore) {
				existingTweet := &models.Tweet{ID: 201, UserID: 202}
				mt.On("GetTweetByTweetID", mock.Anything, int64(201)).Return(existingTweet, nil)
				mt.On("DeleteTweet", mock.Anything, int64(201)).Return(errMockInternal)
			},
			wantedErr: errMockInternal,
			errMsg:    "ツイートの削除に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetStore)
			tt.setupMock(mt)
			svc := NewTweetService(mt)
			ctx := context.Background()

			err := svc.RemoveTweet(ctx, tt.inputTweetID, tt.inputUserID)

			if tt.wantedErr != nil {
				if tt.errMsg != "" {
					assert.ErrorIs(t, err, tt.wantedErr)
					assert.Contains(t, err.Error(), tt.errMsg)
					t.Logf("エラーは %v\n", err)
				}
				assert.ErrorIs(t, err, tt.wantedErr)
			} else {
				require.NoError(t, err)
			}

			mt.AssertExpectations(t)
		})
	}
}
