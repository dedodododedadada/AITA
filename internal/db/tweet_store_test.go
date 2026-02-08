package db

import (
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"aita/internal/pkg/utils"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTweet(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	
	initUser := &models.User{
		Username: "henry",
		Email: "text@example.com",
		PasswordHash:"passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	t.Run("正常にツイートできること", func(t *testing.T) {
		tweet := &models.Tweet{
			UserID: createdUser.ID,
			Content: "これはテスト用のツイート投稿です",
			ImageURL: utils.StringPtr("https://example.com/image.jpg"),
		}
		createdTweet, err := testTweetStore.CreateTweet(ctx, tweet)
		require.NoError(t, err, "CreatedTweetはエラーを返すべきではありません")
		require.NotZero(t, createdTweet.ID, "投稿後、IDが発番されるはずです")
		require.NotZero(t, createdTweet.CreatedAt, "投稿後、作成日時がセットされるはずです")
		require.Equal(t, tweet.UserID, createdTweet.UserID)
		require.Equal(t, tweet.Content, createdTweet.Content)
		require.Equal(t, tweet.ImageURL, createdTweet.ImageURL)
	})
}

func TestCreateTweetWhileError(t *testing.T) {
		testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	
	initUser := &models.User{
		Username: "henry",
		Email: "text@example.com",
		PasswordHash:"passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	
	t.Run("存在しないUserIDNの場合は、ErrUserNotFoundを返すこと", func(t *testing.T) {
		badTweet := &models.Tweet{
			UserID: 99999,
			Content: "存在しないユーザーのツイート",
		}
		createdTweet,err := testTweetStore.CreateTweet(ctx, badTweet)
		require.Error(t, err, "ユーザーが存在しない場合、エラーを返すべきです")
		require.ErrorIs(t, err, models.ErrUserNotFound, "エラーはErrUserNotFoundであるべきです")
		require.Nil(t, createdTweet, "エラー時、生成されたツイートはnilであるべきです")
	})	
	
	t.Run("Contentが制限文字数を超えた場合、ErrValueTooLongを返すこと", func(t *testing.T) {
        longContent := strings.Repeat("a", 3000) 
        badTweet := &models.Tweet{
            UserID:  createdUser.ID,
            Content: longContent,
        }
        _, err := testTweetStore.CreateTweet(ctx, badTweet)
        
        require.Error(t, err)
        require.ErrorIs(t, err, models.ErrValueTooLong, "DBの切捨てエラーが正しくマッピングされること")
    })
	
	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)
		
		tempTweetStore := NewPostgresTweetStore(tempDB)
		tempDB.Close()

		tweet := &models.Tweet{
			UserID: createdUser.ID,
			Content: "これはテスト用のツイート投稿です",
			ImageURL: utils.StringPtr("https://example.com/image.jpg"),
		}
		createdTweet, err := tempTweetStore.CreateTweet(ctx, tweet)
		require.Error(t, err, "内部エラーの場合、エラーを返すべきです")
		assert.Contains(t, err.Error(), "ツイートの挿入に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, createdTweet, "エラー時、生成されたツイートはnilであるべきです")
	})
}
