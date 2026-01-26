package db

import (
	"aita/internal/models"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

	func TestCreateUser(t *testing.T) {
		testContext.CleanupTestDB()
		defer testContext.CleanupTestDB()
		ctx := context.Background()
		req := &models.SignupRequest{
			Username: "henry",
			Email: "text@example.com",
			Password:"a123456",
		}

		createdUser, err := testUserStore.Create(ctx, req)
		require.NoError(t, err, "CreatedUserはエラーを返すべきではありません")
		require.NotNil(t, createdUser, "作成されたユーザーオブジェクトは空であるべきではありません")
		require.Equal(t, req.Username, createdUser.Username, "ユーザー名は一致する必要があります")
		require.Equal(t, req.Email, createdUser.Email, "メールは一致する必要があります")
		require.NotEmpty(t, createdUser.PasswordHash, "パスワードハッシュは空であるべきではありません")
		require.NotEqual(t, req.Password, createdUser.PasswordHash, "パスワードは決して平文で保存してはいけません")
	}

func TestGetByEmail(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	req := &models.SignupRequest{
		Username: "testuser_for_get",
		Email:    "get@example.com",
		Password: "password123",
	}
	createdUser, err := testUserStore.Create(ctx, req)
	require.NoError(t,err)
	t.Run("ユーザーが存在する", func(t *testing.T) {
		foundUser, err := testUserStore.GetByEmail(ctx, "get@example.com")
		require.NoError(t, err, "ユーザーが存在する場合、エラーを返すべきではありません")
		require.NotNil(t,foundUser, "見つかったユーザーは空であってはなりません")
		require.Equal(t,createdUser.ID, foundUser.ID, "見つかったユーザーIDが一致するはずです")
		require.Equal(t,createdUser.Username, foundUser.Username, "見つかったユーザー名が一致するはずです")
	})
	t.Run("ユーザーが存在しない", func(t *testing.T) {
		foundUser, err := testUserStore.GetByEmail(ctx, "nonexistent@example.com")
		require.Error(t,err, "ユーザーが存在しない場合、エラーを返すべきです")
		require.ErrorIs(t,err, ErrNotFound, "エラーはErrNotFound であるべきです")
		require.Nil(t,foundUser, "見つかったユーザーオブジェクトは空であるべきです")
	})
}

func TestGetByID(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	req := &models.SignupRequest{
		Username: "testuser_for_id",
		Email: 	  "getid@example.com",
		Password: "pasword123",
	}
	createdUser, err := testUserStore.Create(ctx, req)
	require.NoError(t, err)
	t.Run("ユーザーが存在する", func(t *testing.T) {
		foundUser, err := testUserStore.GetByID(ctx, createdUser.ID)
		require.NoError(t, err, "アイディが存在する場合、エラーを返すべきではありません")
		require.NotNil(t, foundUser, "見つかったユーザーは空であってはなりません")
		require.Equal(t, createdUser.ID, foundUser.ID, "見つかったユーザーIDが一致するはずです")
		require.Equal(t, createdUser.Username, foundUser.Username, "見つかったユーザー名が一致するはずです")
	})
	t.Run("ユーザーが存在しない", func(t *testing.T) {
		foundUser, err := testUserStore.GetByID(ctx, 99999)
		require.Error(t, err,  "アイディが存在しない場合、エラーを返すべきです")
		require.ErrorIs(t, err, ErrNotFound, "エラーはErrNotFound であるべきです")
		require.Nil(t, foundUser, "見つかったユーザーオブジェクトは空であるべきです")
	})
}
