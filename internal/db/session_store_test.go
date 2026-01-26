package db

import (
	"aita/internal/models"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	userReq := &models.SignupRequest{
		Username: "token_test_user",
		Email:    "token_test@example.com",
		Password: "password456",
	}
	user, err := testUserStore.Create(ctx, userReq)
	require.NoError(t, err)
	duration := 24 *time.Hour
	createdToken, createdSession, err := testSessionStore.Create(ctx, user.ID, duration)
	require.NoError(t, err, "CreatedTokenはエラーを返すべきではありません")
	require.NotEmpty(t, createdToken, "作成されたトークンは空であるべきではありません")
	require.NotNil(t, createdSession, "作成されたセッションは空であるべきではありません")
	require.Equal(t, user.ID, createdSession.UserID, "ユーザーIDは一致する必要があります")
	require.NotEmpty(t, createdSession.TokenHash, "TokenHashは空であるべきではありません")
	require.WithinDuration(t, time.Now().Add(duration), createdSession.ExpiresAt, 2*time.Second,
						  "期待される有効期限が、現在の時刻+durationから2秒以内であることを検証")
}

func TestGetByToken(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	userReq := &models.SignupRequest{
		Username: "test_get_token",
		Email:    "get_token@example.com",
		Password: "password789" , 
	}
	user, err := testUserStore.Create(ctx, userReq)
	require.NoError(t, err)
	duration := 1 * time.Hour
	rawToken, createdSession, _ := testSessionStore.Create(ctx, user.ID, duration)
	t.Run("有効なトークンで取得", func(t *testing.T) {
		foundSession, err := testSessionStore.GetByToken(ctx, rawToken)
		require.NoError(t, err, "トークンが有効な場合は、エラーを返すべきではありません")
		require.NotNil(t, foundSession, "見つかったセッションは空であってはなりません")
		require.Equal(t, createdSession.ID, foundSession.ID, "見つかったセッションIDが一致するはずです")
		require.Equal(t, user.ID, foundSession.UserID, "見つかったユーザーIDが一致するはずです")
	})
	t.Run("トークン", func(t *testing.T){
		invalidSession, err := testSessionStore.GetByToken(ctx, "non-existent-token")
		require.Error(t, err, "トークンがない場合、エラーを返すべきです")
		require.Nil(t, invalidSession, "トークンがない場合、セッションオブジェクトは空であるべきです")
		require.ErrorIs(t, err, ErrNotFound, "エラーはErrNotFound であるべきです")
	})
}