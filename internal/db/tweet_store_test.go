package db

import (
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTweet(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	
	req := &models.SignupRequest{
			Username: "henry",
			Email: "text@example.com",
			Password:"a123456",
		}

	createdUser, err := testUserStore.Create(ctx, req)
	require.NoError(t, err)
	t.Run("正常にツイートできること", func(t *testing.T) {
		tweet := &models.Tweet{
			UserID: createdUser.ID,
			Content: "これはテスト用のツイート投稿です",
			ImageURL: utils.StringPtr("https://example.com/image.jpg"),
		}
		err = testTweetStore.CreateTweet(ctx, tweet)
		require.NoError(t, err, "CreatedTweetはエラーを返すべきではありません")
		require.NotZero(t, tweet.UserID, "投稿後、IDが発番されるはずです")
		require.NotZero(t, tweet.CreatedAt, "投稿後、作成日時がセットされるはずです")
		// re-confirmed by getTweetById
	})
	t.Run("存在しないUserIDNの場合は、ErrUserNotFoundを返すこと", func(t *testing.T) {
		badTweet := &models.Tweet{
			UserID: 99999,
			Content: "存在しないユーザーのツイート",
		}
		err = testTweetStore.CreateTweet(ctx, badTweet)
		require.Error(t, err, "ユーザーが存在しない場合、エラーを返すべきです")
		require.ErrorIs(t, err, models.ErrUserNotFound, "エラーはErrUserNotFoundであるべきです")
		require.Zero(t, badTweet.ID, "エラー時はIDが生成されないはずです")
		require.Zero(t, badTweet.CreatedAt, "エラー時、タイムスタンプが生成されないはずです")
	})
	
}