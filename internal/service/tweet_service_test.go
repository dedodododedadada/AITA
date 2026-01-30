package service

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostTweet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{
		name			string
		userID          int64
		inputBody       *models.CreateTweetRequest
		setupMock       func(mt *MockTweetStore)
		wantedErr       error
	}{
		{
			name: "発送成功",
			userID: 101,
			inputBody: &models.CreateTweetRequest{
				Content: "Hello world",
				ImageURL: utils.StringPtr("https://example.com/mock.jpg"),
			},
			setupMock: func(mt *MockTweetStore) {
				mt.On("CreateTweet", mock.Anything, mock.MatchedBy(func(t *models.Tweet) bool {
					return t.UserID == 101 && t.Content == "Hello world"
				})).Return(nil)
			},
			wantedErr: nil,
		},
		{
			name: "パラメーターエラー、コンテントは空である",
			userID: 101,
			inputBody:&models.CreateTweetRequest{
				Content: "",
			},
			setupMock: func(mt *MockTweetStore){},
			wantedErr: models.ErrContentEmpty,
		},
		{
			name: "データベースエラー,ユーザーが存在しない場合",
			userID: 99999,
			inputBody: &models.CreateTweetRequest{
				Content: "Hello world",
				ImageURL: utils.StringPtr("https://example.com/mock.jpg"),
			},
			setupMock: func(mt *MockTweetStore) {
				mt.On("CreateTweet", mock.Anything, mock.MatchedBy(func(t *models.Tweet) bool {
					if t.UserID != 99999 || t.Content != "Hello world" {
						return false
					}	
					if t.ImageURL == nil || *t.ImageURL != "https://example.com/mock.jpg" {
						return false
					}
					return true
				})).Return(fmt.Errorf("ツイートの挿入に失敗しました: %w", models.ErrUserNotFound))
			},
			wantedErr: models.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(MockTweetStore)
			tt.setupMock(mt)
			svc := NewTweetService(mt)
			ctx := context.Background()
			res, err := svc.PostTweet(ctx, tt.userID, tt.inputBody)

			if tt.wantedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantedErr,"期待されるエラータイプが一致します")
				require.Nil(t, res, "エラーが発生した場合は、レスポンスはnilであるべきです")
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, tt.inputBody.Content, res.Content)
				require.Equal(t, tt.inputBody.ImageURL, res.ImageURL)
			}

			mt.AssertExpectations(t)
		})
	}
}