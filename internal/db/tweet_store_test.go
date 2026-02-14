package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"aita/internal/pkg/utils"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTweet(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	t.Run("正常にツイートできること", func(t *testing.T) {
		tweet := &models.Tweet{
			UserID:   createdUser.ID,
			Content:  "これはテスト用のツイート投稿です",
			ImageURL: utils.StringPtr("https://example.com/image.jpg"),
		}
		createdTweet, err := testTweetStore.CreateTweet(ctx, tweet)
		require.NoError(t, err, "CreatedTweetはエラーを返すべきではありません")
		require.NotZero(t, createdTweet.ID, "投稿後、IDが発番されるはずです")
		require.NotZero(t, createdTweet.CreatedAt, "投稿後、作成日時がセットされるはずです")
		require.NotZero(t, createdTweet.CreatedAt)
    	require.NotZero(t, createdTweet.UpdatedAt)
		assert.Equal(t, tweet.UserID, createdTweet.UserID)
		assert.Equal(t, tweet.Content, createdTweet.Content)
		assert.Equal(t, tweet.ImageURL, createdTweet.ImageURL)
		assert.WithinDuration(t, createdTweet.CreatedAt, createdTweet.UpdatedAt, 1*time.Second)
		assert.Equal(t, time.UTC, createdTweet.CreatedAt.Location())
		assert.Equal(t, time.UTC, createdTweet.UpdatedAt.Location())
		assert.WithinDuration(t, time.Now(), createdTweet.CreatedAt, 10*time.Second)
		assert.Equal(t, false, createdTweet.IsEdited)
	})
}

func TestCreateTweetWhileError(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("存在しないUserIDNの場合は、ErrUserNotFoundを返すこと", func(t *testing.T) {
		badTweet := &models.Tweet{
			UserID:  99999,
			Content: "存在しないユーザーのツイート",
		}
		createdTweet, err := testTweetStore.CreateTweet(ctx, badTweet)
		require.Error(t, err, "ユーザーが存在しない場合、エラーを返すべきです")
		require.ErrorIs(t, err, errcode.ErrUserNotFound, "エラーはErrUserNotFoundであるべきです")
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
		require.ErrorIs(t, err, errcode.ErrValueTooLong, "DBの切捨てエラーが正しくマッピングされること")
	})

	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempTweetStore := NewPostgresTweetStore(tempDB)
		tempDB.Close()

		tweet := &models.Tweet{
			UserID:   createdUser.ID,
			Content:  "これはテスト用のツイート投稿です",
			ImageURL: utils.StringPtr("https://example.com/image.jpg"),
		}
		createdTweet, err := tempTweetStore.CreateTweet(ctx, tweet)
		require.Error(t, err, "内部エラーの場合、エラーを返すべきです")
		assert.Contains(t, err.Error(), "ツイートの挿入に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, createdTweet, "エラー時、生成されたツイートはnilであるべきです")
	})
}

func TestGetTweetByTweetID(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initTweet := &models.Tweet{
		UserID:   createdUser.ID,
		Content:  "これはテスト用のツイート投稿です",
		ImageURL: utils.StringPtr("https://example.com/image.jpg"),
	}

	createdTweet, err := testTweetStore.CreateTweet(ctx, initTweet)
	require.NoError(t, err)

	t.Run("正常系:ツイートをTweetIDで取得に成功したこと", func(t *testing.T) {
		foundTweet, err := testTweetStore.GetTweetByTweetID(ctx, createdTweet.ID)
		require.NoError(t, err)
		require.NotNil(t, foundTweet)
		assert.Equal(t, createdTweet.UserID, foundTweet.UserID)
		assert.Equal(t, createdTweet.Content, foundTweet.Content)
		assert.Equal(t, createdTweet.ImageURL, foundTweet.ImageURL)
		assert.Equal(t, createdTweet.CreatedAt, foundTweet.CreatedAt)
		assert.Equal(t, createdTweet.UpdatedAt, foundTweet.UpdatedAt)
		assert.Equal(t, time.UTC, createdTweet.CreatedAt.Location())
		assert.Equal(t, time.UTC, createdTweet.UpdatedAt.Location())
		assert.Equal(t, createdTweet.IsEdited, foundTweet.IsEdited)
	})
}

func TestGetTweetByTweetIDWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initTweet := &models.Tweet{
		UserID:   createdUser.ID,
		Content:  "これはテスト用のツイート投稿です",
		ImageURL: utils.StringPtr("https://example.com/image.jpg"),
	}

	createdTweet, err := testTweetStore.CreateTweet(ctx, initTweet)
	require.NoError(t, err)

	t.Run("異常系:TweetIDがないこと", func(t *testing.T) {
		foundTweet, err := testTweetStore.GetTweetByTweetID(ctx, int64(999))
		assert.ErrorIs(t, err, errcode.ErrTweetNotFound)
		assert.Nil(t, foundTweet)
	})

	t.Run("異常系:データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)
		tempTweetStore := NewPostgresTweetStore(tempDB)
		tempDB.Close()
		foundTweet, err := tempTweetStore.GetTweetByTweetID(ctx, createdTweet.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ツイートの取得に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, foundTweet)
	})
}

func TestUpdateContent(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initTweet := &models.Tweet{
		UserID:   createdUser.ID,
		Content:  "これはテスト用のツイート投稿です",
		ImageURL: utils.StringPtr("https://example.com/image.jpg"),
	}

	createdTweet, err := testTweetStore.CreateTweet(ctx, initTweet)
	require.NoError(t, err)

	t.Run("正常系:ツイートの更新に成功したこと", func(t *testing.T) {
		newContent := "Contentが更新しました"
		updatedTweet, err := testTweetStore.UpdateContent(ctx, newContent, createdTweet.ID)
		require.NoError(t, err)
		require.NotNil(t, updatedTweet)
		assert.Equal(t, createdTweet.ID, updatedTweet.ID)
		assert.Equal(t, createdTweet.UserID, updatedTweet.UserID)
		assert.Equal(t, createdTweet.CreatedAt, updatedTweet.CreatedAt)
		assert.Equal(t, newContent, updatedTweet.Content)
		assert.WithinDuration(t, time.Now(), updatedTweet.UpdatedAt, 2*time.Second)
		assert.Equal(t, true, updatedTweet.IsEdited)
	})
}

func TestUpdateContentWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initTweet := &models.Tweet{
		UserID:   createdUser.ID,
		Content:  "これはテスト用のツイート投稿です",
		ImageURL: utils.StringPtr("https://example.com/image.jpg"),
	}

	createdTweet, err := testTweetStore.CreateTweet(ctx, initTweet)
	require.NoError(t, err)

	t.Run("異常系:contentが長すぎ", func(t *testing.T) {
		newContent := strings.Repeat("a", 1001)
		updatedTweet, err := testTweetStore.UpdateContent(ctx, newContent, createdTweet.ID)
		assert.ErrorIs(t, err, errcode.ErrValueTooLong)
		assert.Nil(t, updatedTweet)
	})

	t.Run("異常系:TweetIDがないこと", func(t *testing.T) {
		newContent := "Contentが更新しました"
		updatedTweet, err := testTweetStore.UpdateContent(ctx, newContent, int64(999))
		assert.ErrorIs(t, err, errcode.ErrTweetNotFound)
		assert.Nil(t, updatedTweet)
	})

	t.Run("異常系:データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)
		tempTweetStore := NewPostgresTweetStore(tempDB)
		tempDB.Close()

		newContent := "Contentが更新しました"
		updatedTweet, err := tempTweetStore.UpdateContent(ctx, newContent, createdTweet.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ツイートの更新に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, updatedTweet)
	})

}

func TestDeleteTweet(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initTweet := &models.Tweet{
		UserID:   createdUser.ID,
		Content:  "これはテスト用のツイート投稿です",
		ImageURL: utils.StringPtr("https://example.com/image.jpg"),
	}
	createdTweet, err := testTweetStore.CreateTweet(ctx, initTweet)

	t.Run("正常系: ツイートの削除に成功したこと", func(t *testing.T) {
		err = testTweetStore.DeleteTweet(ctx, createdTweet.ID)
		require.NoError(t, err)
		res, err := testTweetStore.GetTweetByTweetID(ctx, createdTweet.ID)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestDeleteTweetWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initTweet := &models.Tweet{
		UserID:   createdUser.ID,
		Content:  "これはテスト用のツイート投稿です",
		ImageURL: utils.StringPtr("https://example.com/image.jpg"),
	}
	createdTweet, err := testTweetStore.CreateTweet(ctx, initTweet)

	t.Run("異常系: TweetIDがないこと", func(t *testing.T) {
		err = testTweetStore.DeleteTweet(ctx, int64(999))
		assert.ErrorIs(t, err, errcode.ErrTweetNotFound)
	})

	t.Run("異常系:データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		temp, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)
		tempTweetStore := NewPostgresTweetStore(temp)
		temp.Close()

		err = tempTweetStore.DeleteTweet(ctx, createdTweet.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ツイートの削除に失敗しました")
		t.Logf("エラーは: %v\n", err)
	})

}
