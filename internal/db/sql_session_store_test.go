package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSession(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "token_test_user",
		Email:        "token_test@example.com",
		PasswordHash: "hashed_password",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("正常にセッションできること", func(t *testing.T) {
		mockHash := "mock_hash_value_123"
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		sessionToCreate := &models.Session{
			UserID:    createdUser.ID,
			TokenHash: mockHash,
			ExpiresAt: expiresAt,
		}
		createdSession, err := testSessionStore.Create(ctx, sessionToCreate)

		require.NoError(t, err, "Sessionの作成に失敗してはいけません")
		require.NotNil(t, createdSession)
		assert.Equal(t, createdUser.ID, createdSession.UserID)
		assert.Equal(t, mockHash, createdSession.TokenHash)
		assert.WithinDuration(t, expiresAt, createdSession.ExpiresAt, time.Second)
		assert.NotZero(t, createdSession.CreatedAt)
	})

}

func TestCreateSessionWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "token_test_user",
		Email:        "token_test@example.com",
		PasswordHash: "hashed_password",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("23503: ユーザーがない", func(t *testing.T) {
		mockHash := "mock_hash_value_123"
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		sessionToCreate := &models.Session{
			UserID:    int64(999999),
			TokenHash: mockHash,
			ExpiresAt: expiresAt,
		}
		createdSession, err := testSessionStore.Create(ctx, sessionToCreate)

		assert.ErrorIs(t, err, errcode.ErrUserNotFound)
		assert.Nil(t, createdSession)
	})

	t.Run("23505: トークンハッシュの重複 (Unique Violation)", func(t *testing.T) {
		mockHash := "duplicate_hash"
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		initSession := &models.Session{
			UserID:    createdUser.ID,
			TokenHash: mockHash,
			ExpiresAt: expiresAt,
		}
		_, err := testSessionStore.Create(ctx, initSession)
		require.NoError(t, err)

		newSession := &models.Session{
			UserID:    createdUser.ID,
			TokenHash: mockHash,
			ExpiresAt: expiresAt,
		}
		createdSession, err := testSessionStore.Create(ctx, newSession)

		assert.ErrorIs(t, err, errcode.ErrTokenConflict)
		assert.Nil(t, createdSession)
	})

	t.Run("22001: tokenhash列長超過", func(t *testing.T) {
		mockHash := strings.Repeat("a", 256)
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		sessionToCreate := &models.Session{
			UserID:    createdUser.ID,
			TokenHash: mockHash,
			ExpiresAt: expiresAt,
		}
		createdSession, err := testSessionStore.Create(ctx, sessionToCreate)

		assert.ErrorIs(t, err, errcode.ErrValueTooLong)
		assert.Nil(t, createdSession)
	})

	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempSessionStore := NewPostgresSessionStore(tempDB)
		tempDB.Close()

		mockHash := "mock_hash_value_123"
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		sessionToCreate := &models.Session{
			UserID:    createdUser.ID,
			TokenHash: mockHash,
			ExpiresAt: expiresAt,
		}
		createdUser, err := tempSessionStore.Create(ctx, sessionToCreate)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "セッションの生成に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, createdUser)
	})
}

func TestGetByHash(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	newUser := &models.User{
		Username:     "test_get_token",
		Email:        "get_token@example.com",
		PasswordHash: "hash",
	}
	createdUser, err := testUserStore.Create(ctx, newUser)
	require.NoError(t, err)

	targetHash := "target_secret_hash"
	initSession := &models.Session{
		UserID:    createdUser.ID,
		TokenHash: targetHash,
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}
	createdSession, _ := testSessionStore.Create(ctx, initSession)
	require.NoError(t, err)

	t.Run("存在するHashで取得", func(t *testing.T) {
		foundSession, err := testSessionStore.GetByHash(ctx, targetHash)

		require.NoError(t, err)
		require.NotNil(t, foundSession)
		assert.Equal(t, createdSession.ID, foundSession.ID)
		assert.Equal(t, createdSession.TokenHash, foundSession.TokenHash)
		assert.Equal(t, createdSession.UserID, foundSession.UserID)
		assert.Equal(t, createdSession.ExpiresAt, foundSession.ExpiresAt)
		assert.Equal(t, createdSession.CreatedAt, foundSession.CreatedAt)
	})
}

func TestGetByHashWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	newUser := &models.User{
		Username:     "test_get_token",
		Email:        "get_token@example.com",
		PasswordHash: "hash",
	}
	createdUser, err := testUserStore.Create(ctx, newUser)
	require.NoError(t, err)

	targetHash := "target_secret_hash"
	initSession := &models.Session{
		UserID:    createdUser.ID,
		TokenHash: targetHash,
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}
	_, err = testSessionStore.Create(ctx, initSession)
	require.NoError(t, err)

	t.Run("tokenhashが存在しない", func(t *testing.T) {
		wrongHash := "wrong_test_toke "
		foundSession, err := testSessionStore.GetByHash(ctx, wrongHash)

		assert.ErrorIs(t, err, errcode.ErrSessionNotFound)
		assert.Nil(t, foundSession)
	})

	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempSessionStore := NewPostgresSessionStore(tempDB)
		tempDB.Close()

		unfoundSession, err := tempSessionStore.GetByHash(ctx, targetHash)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "セッション取得に失敗")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, unfoundSession)
	})
}

func TestUpdateExpiredAt(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "update_test_user",
		Email:        "update_test@example.com",
		PasswordHash: "hashed_password",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initialSession := &models.Session{
		UserID:    createdUser.ID,
		TokenHash: "initial_hash",
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}
	createdSession, err := testSessionStore.Create(ctx, initialSession)
	require.NoError(t, err)

	t.Run("正常に有効期限を更新できること", func(t *testing.T) {
		newExpiry := time.Now().Add(48 * time.Hour).UTC()

		err := testSessionStore.UpdateExpiresAt(ctx, newExpiry, createdSession.ID)

		assert.NoError(t, err)

		updatedSession, err := testSessionStore.GetByHash(ctx, "initial_hash")
		require.NoError(t, err)
		assert.WithinDuration(t, newExpiry, updatedSession.ExpiresAt, time.Second)
	})
}

func TestUpdateExpiredAtWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "update_test_user",
		Email:        "update_test@example.com",
		PasswordHash: "hashed_password",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	initialSession := &models.Session{
		UserID:    createdUser.ID,
		TokenHash: "initial_hash",
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}
	createdSession, err := testSessionStore.Create(ctx, initialSession)
	require.NoError(t, err)

	t.Run("存在しないIDを指定した場合、ErrSessionNotFoundを返すこと", func(t *testing.T) {
		nonExistentID := int64(999999)
		newExpiry := time.Now().Add(24 * time.Hour).UTC()

		err := testSessionStore.UpdateExpiresAt(ctx, newExpiry, nonExistentID)
		assert.ErrorIs(t, err, errcode.ErrSessionNotFound)
	})

	t.Run("データベース切断時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, _ := testutils.OpenDB(testContext.DSN)
		tempStore := NewPostgresSessionStore(tempDB)
		tempDB.Close()

		err := tempStore.UpdateExpiresAt(ctx, time.Now(), createdSession.ID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "セッション期限の更新に失敗しました")
		t.Logf("期待通りキャッチされたエラー: %v", err)
	})
}

func TestDeleteBySessionID(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()

	initUser := &models.User{
		Username:     "update_test_user",
		Email:        "update_test@example.com",
		PasswordHash: "hashed_password",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	hash := "initial_hash"
	initialSession := &models.Session{
		UserID:    createdUser.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}
	createdSession, err := testSessionStore.Create(ctx, initialSession)
	require.NoError(t, err)
	require.NotNil(t, createdSession)

	t.Run("正常にハッシュで削除できること", func(t *testing.T) {
		err = testSessionStore.DeleteBySessionID(ctx, createdSession.ID)
		require.NoError(t, err)
		test, err := testSessionStore.GetByHash(ctx, hash)
		require.ErrorIs(t, err, errcode.ErrSessionNotFound)
		require.Nil(t, test)
	})
	t.Run("データベース切断時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, _ := testutils.OpenDB(testContext.DSN)
		tempStore := NewPostgresSessionStore(tempDB)
		tempDB.Close()

		err := tempStore.DeleteBySessionID(ctx, createdSession.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "セッションの削除に失敗しました")
		t.Logf("期待通りキャッチされたエラー: %v", err)
	})

	t.Run("存在しないIDを指定した場合、ErrSessionNotFoundを返すこと", func(t *testing.T) {
		wrongID:= int64(999)
		err := testSessionStore.DeleteBySessionID(ctx, wrongID)
		assert.ErrorIs(t, err, errcode.ErrSessionNotFound)
	})
}

func TestDeleteAllByUserID(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "update_test_user",
		Email:        "update_test@example.com",
		PasswordHash: "hashed_password",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	hashList := []string{"token_hash_1", "token_hash_2", "token_hash_3"}
	for i := range hashList {
		initSession := &models.Session{
			UserID:    createdUser.ID,
			TokenHash: hashList[i],
			ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
		}
		createdSession, err := testSessionStore.Create(ctx, initSession)
		require.NoError(t, err)
		require.NotNil(t, createdSession)
	}

	t.Run("正常にユーザーIDで削除できること", func(t *testing.T) {
		err = testSessionStore.DeleteAllByUserID(ctx, createdUser.ID)
		require.NoError(t, err)
		for i := range hashList {
			createdSession, err := testSessionStore.GetByHash(ctx, hashList[i])
			require.ErrorIs(t, err, errcode.ErrSessionNotFound)
			require.Nil(t, createdSession)
		}
	})

	t.Run("データベース切断時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, _ := testutils.OpenDB(testContext.DSN)
		tempStore := NewPostgresSessionStore(tempDB)
		tempDB.Close()

		err := tempStore.DeleteAllByUserID(ctx, createdUser.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ユーザーの全セッション削除に失敗しました")
		t.Logf("期待通りキャッチされたエラー: %v", err)
	})
}
