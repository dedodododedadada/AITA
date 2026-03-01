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

func TestCreate(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "hashedpassword",
	}

	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err, "CreatedUserはエラーを返すべきではありません")
	require.NotNil(t, createdUser, "作成されたユーザーオブジェクトは空であるべきではありません")
	assert.Equal(t, initUser.Username, createdUser.Username, "ユーザー名は一致するべきです")
	assert.Equal(t, initUser.Email, createdUser.Email, "メールは一致する必要があります")
	assert.Equal(t, initUser.PasswordHash, createdUser.PasswordHash, "パスワードハッシュは空であるべきではありません")
	assert.Zero(t, createdUser.FollowerCount)
	assert.Zero(t, createdUser.FollowingCount)
	assert.NotZero(t, createdUser.ID, "IDは自動生成されるべきです")
	assert.NotZero(t, createdUser.CreatedAt, "CreatedAtは自動生成されるべきです")
	assert.Equal(t, time.UTC, createdUser.CreatedAt.Location(), "CreatedAtはUTCであるべきです")
}

func TestCreateWhileConflict(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "henry",
		Email:        "text@example.com",
		PasswordHash: "hashedpassword",
	}
	_, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("23505: ユーザー名の競合", func(t *testing.T) {
		newUser := &models.User{
			Username:     "henry",
			Email:        "other@example.com",
			PasswordHash: "secretpassword",
		}
		createdUser, err := testUserStore.Create(ctx, newUser)

		assert.ErrorIs(t, err, errcode.ErrUsernameConflict, "エラーはErrUsernameConflictであるべきです")
		assert.Nil(t, createdUser, "エラー時、生成されたユーザーはnilであるべきです")
	})

	t.Run("23505: メールアドレスの競合", func(t *testing.T) {
		newUser := &models.User{
			Username:     "Test_name",
			Email:        "text@example.com",
			PasswordHash: "secretpassword",
		}
		createdUser, err := testUserStore.Create(ctx, newUser)
		assert.ErrorIs(t, err, errcode.ErrEmailConflict, "エラーはErrEmailConflictであるべきです")
		assert.Nil(t, createdUser, "エラー時、生成されたユーザーはnilであるべきです")
	})
}

func TestCreateWhileAnotherErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	t.Run("22001: 文字列長超過", func(t *testing.T) {
		newUser := &models.User{
			Username:     strings.Repeat("u", 51),
			Email:        "long@example.com",
			PasswordHash: "sercretpassword",
		}
		createdUser, err := testUserStore.Create(ctx, newUser)
		assert.ErrorIs(t, err, errcode.ErrValueTooLong)
		assert.Nil(t, createdUser)

	})

	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempUserStore := NewPostgresUserStore(tempDB)
		tempDB.Close()

		newUser := &models.User{
			Username:     "henry",
			Email:        "text@example.com",
			PasswordHash: "hashedpassword",
		}
		createdUser, err := tempUserStore.Create(ctx, newUser)
		require.Error(t, err, "内部エラーの場合、エラーを返すべきです")
		assert.Contains(t, err.Error(), "ユーザーの生成に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, createdUser, "エラー時、生成されたユーザーはnilであるべきです")
	})
}

func TestGetByEmail(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "henry",
		Email:        "get@example.com",
		PasswordHash: "hashedpassword",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	t.Run("ユーザーが存在する", func(t *testing.T) {
		foundUser, err := testUserStore.GetByEmail(ctx, "get@example.com")
		require.NoError(t, err, "ユーザーが存在する場合、エラーを返すべきではありません")
		require.NotNil(t, foundUser, "見つかったユーザーは空であってはなりません")
		assert.Equal(t, createdUser.ID, foundUser.ID, "見つかったユーザーIDが一致するはずです")
		assert.Equal(t, createdUser.Username, foundUser.Username, "見つかったユーザー名が一致するはずです")
		assert.Equal(t, createdUser.Email, foundUser.Email, "見つかったメールアドレスが一致するはずです")
		assert.Equal(t, createdUser.PasswordHash, foundUser.PasswordHash, "見つかったPasswordHashが一致するはずです")
		assert.True(t, createdUser.CreatedAt.Equal(foundUser.CreatedAt), "CreatedAtは一致するべきです")
		assert.Equal(t, createdUser.FollowerCount, foundUser.FollowerCount)
		assert.Equal(t, createdUser.FollowingCount, foundUser.FollowingCount)
	})
}

func TestGetByEmailWhileError(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "henry",
		Email:        "get@example.com",
		PasswordHash: "hashedpassword",
	}
	_, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("ユーザーが存在しない", func(t *testing.T) {
		unfoundUser, err := testUserStore.GetByEmail(ctx, "wrong@example.com")
		assert.ErrorIs(t, err, errcode.ErrUserNotFound)
		assert.Nil(t, unfoundUser)
	})

	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempUserStore := NewPostgresUserStore(tempDB)
		tempDB.Close()

		unfoundUser, err := tempUserStore.GetByEmail(ctx, "get@example.com")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "emailによるユーザー取得に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, unfoundUser)
	})
}

func TestGetByID(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "testuser_for_id",
		Email:        "getid@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	foundUser, err := testUserStore.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "アイディが存在する場合、エラーを返すべきではありません")
	require.NotNil(t, foundUser, "見つかったユーザーは空であるべきではありません")
	assert.Equal(t, createdUser.ID, foundUser.ID, "見つかったユーザーIDが一致するべきです")
	assert.Equal(t, createdUser.Username, foundUser.Username, "見つかったユーザー名が一致するべきです")
	assert.Equal(t, createdUser.Email, foundUser.Email, "見つかったメールアドレスが一致するべきです")
	assert.Equal(t, createdUser.PasswordHash, foundUser.PasswordHash, "見つかったPasswordHashが一致するべきです")
	assert.True(t, createdUser.CreatedAt.Equal(foundUser.CreatedAt), "CreatedAtは一致するべきです")
	assert.Equal(t, createdUser.FollowerCount, foundUser.FollowerCount)
	assert.Equal(t, createdUser.FollowingCount, foundUser.FollowingCount)
}

func TestGetByIDWhlieError(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	initUser := &models.User{
		Username:     "testuser_for_id",
		Email:        "getid@example.com",
		PasswordHash: "passwordHash",
	}
	_, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)
	t.Run("ユーザーが存在しない", func(t *testing.T) {
		unfoundUser, err := testUserStore.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, errcode.ErrUserNotFound, "エラーはErrNotFound であるべきです")
		assert.Nil(t, unfoundUser, "見つかったユーザーオブジェクトは空であるべきです")
	})

	t.Run("データベース切断時の時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempUserStore := NewPostgresUserStore(tempDB)
		tempDB.Close()

		unfoundUser, err := tempUserStore.GetByID(ctx, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "IDによるユーザー取得に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, unfoundUser)
	})
}

func TestIncreFollowerCount(t *testing.T) {
    testContext.CleanupTestDB()
    defer testContext.CleanupTestDB()
   	ctx := context.Background()

    
    initUser := &models.User{
        Username:     "follower_test_user",
        Email:        "follower@example.com",
        PasswordHash: "passwordHash",
    }
    createdUser, err := testUserStore.Create(ctx, initUser)
    require.NoError(t, err)


    t.Run("正常系：フォロワー数のインクリメント成功", func(t *testing.T) {
       	tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
		require.NoError(t, err)
		ctx = injectTx(ctx, tx)
		
		delta := int64(1)
        err = testUserStore.IncreaseFollowerCount(ctx, createdUser.ID, delta)
        require.NoError(t, err)
        
        err = tx.Commit()
        require.NoError(t, err)
        
        updatedUser, err := testUserStore.GetByID(ctx, createdUser.ID)
        require.NoError(t, err)
        assert.Equal(t, int64(1), updatedUser.FollowerCount)
    })

    t.Run("正常系：フォロワー数のデクリメント成功", func(t *testing.T) {
        tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
        require.NoError(t, err)
		ctx = injectTx(ctx, tx)
        
        delta := int64(-1)
        err = testUserStore.IncreaseFollowerCount(ctx, createdUser.ID, delta)
        require.NoError(t, err)
        
        err = tx.Commit()
        require.NoError(t, err)
        
        updatedUser, err := testUserStore.GetByID(ctx, createdUser.ID)
        require.NoError(t, err)
        assert.Equal(t, int64(0), updatedUser.FollowerCount)
    })

    t.Run("正常系：負の更新による下限値（0）の維持検証", func(t *testing.T) {
        tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
        require.NoError(t, err)
		ctx = injectTx(ctx, tx)
        
        delta := int64(-100)
        err = testUserStore.IncreaseFollowerCount(ctx, createdUser.ID, delta)
        require.NoError(t, err)
        
        err = tx.Commit()
        require.NoError(t, err)

        updatedUser, err := testUserStore.GetByID(ctx, createdUser.ID)
        require.NoError(t, err)
        assert.Equal(t, int64(0), updatedUser.FollowerCount)
    })
}
func TestIncreFollowerCountWhileErr(t *testing.T) {
    testContext.CleanupTestDB()
    defer testContext.CleanupTestDB()
    ctx := context.Background()
    
    initUser := &models.User{
        Username:     "follower_err_user",
        Email:        "follower_err@example.com",
        PasswordHash: "passwordHash",
    }
    createdUser, err := testUserStore.Create(ctx, initUser)
    require.NoError(t, err)

    t.Run("異常系：存在しないユーザーID", func(t *testing.T) {
        tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback() 
        require.NoError(t, err)
		ctx = injectTx(ctx, tx)

        nonExistentID := int64(99999)
        err = testUserStore.IncreaseFollowerCount(ctx, nonExistentID, 1)
        
        assert.ErrorIs(t, err, errcode.ErrUserNotFound)
    })

    t.Run("異常系：ロールバックでデータが戻ること", func(t *testing.T) {
        tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
        require.NoError(t, err)
		ctx = injectTx(ctx, tx)
        
        err = testUserStore.IncreaseFollowerCount(ctx, createdUser.ID, 50)
        require.NoError(t, err)
        err = tx.Rollback()
        require.NoError(t, err)
        
        finalUser, err := testUserStore.GetByID(ctx, createdUser.ID)
        require.NoError(t, err)
        assert.Equal(t, int64(0), finalUser.FollowerCount)
    })

    t.Run("異常系：DB切断時のラップされたエラー", func(t *testing.T) {
        tempDB, err := testutils.OpenDB(testContext.DSN)
        require.NoError(t, err)
        tempUserStore := NewPostgresUserStore(tempDB)
        tempDB.Close()

        err = tempUserStore.IncreaseFollowerCount(ctx,  createdUser.ID, 1)
        
        require.Error(t, err)
        assert.Contains(t, err.Error(), "follower countsの更新に失敗しました")
    })
}

func TestIncreFollowingCount(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	
	initUser := &models.User{
		Username:     "testuser_for_id",
		Email:        "getid@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("正常系：更新成功(follow)",func(t *testing.T) {
		tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
    	require.NoError(t, err)
		ctx = injectTx(ctx, tx)
		
		delta := int64(1)
		err = testUserStore.IncreaseFollowingCount(ctx, createdUser.ID, delta)
		
		require.NoError(t, err)
		err = tx.Commit()
		require.NoError(t,err)
		updatedUser, err := testUserStore.GetByID(ctx, createdUser.ID)
    	require.NoError(t, err)
    	assert.Equal(t, int64(1), updatedUser.FollowingCount)
	})
	t.Run("正常系：更新成功(UnFollow)",func(t *testing.T) {
		tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
    	require.NoError(t, err)
		ctx = injectTx(ctx, tx)

		delta := int64(-1)
		err = testUserStore.IncreaseFollowingCount(ctx, createdUser.ID, delta)
		
		require.NoError(t, err)
		err = tx.Commit()
		require.NoError(t,err)
		updatedUser, err := testUserStore.GetByID(ctx, createdUser.ID)
    	require.NoError(t, err)
    	assert.Equal(t, int64(0), updatedUser.FollowingCount)
	})

	t.Run("正常系：負の更新による0下限の維持", func(t *testing.T) {
		tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
    	require.NoError(t, err)
		ctx = injectTx(ctx, tx)

		delta := int64(1)
		err = testUserStore.IncreaseFollowingCount(ctx, createdUser.ID, delta)
		require.NoError(t, err)
		
		delta = int64(-9)
		err = testUserStore.IncreaseFollowingCount(ctx, createdUser.ID, delta)
		err = tx.Commit()
		assert.NoError(t, err)
		updatedUser, err := testUserStore.GetByID(ctx, createdUser.ID)
    	require.NoError(t, err)
    	assert.Equal(t, int64(0), updatedUser.FollowingCount)
	})
}

func TestIncreFollowingCountWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	defer testContext.CleanupTestDB()
	ctx := context.Background()
	
	initUser := &models.User{
		Username:     "testuser_for_id",
		Email:        "getid@example.com",
		PasswordHash: "passwordHash",
	}
	createdUser, err := testUserStore.Create(ctx, initUser)
	require.NoError(t, err)

	t.Run("異常系：userなし",func(t *testing.T) {
		tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
    	require.NoError(t, err)
		ctx = injectTx(ctx, tx)

		delta := int64(1)
		err = testUserStore.IncreaseFollowingCount(ctx, int64(101), delta)
		
		assert.ErrorIs(t, err, errcode.ErrUserNotFound)
	})

	t.Run("異常系：ロールバックでデータが戻ること", func(t *testing.T) {
		tx, err := testContext.TestDB.BeginTxx(ctx, nil)
		defer tx.Rollback()
		require.NoError(t, err)
		ctx = injectTx(ctx, tx)

		err = testUserStore.IncreaseFollowingCount(ctx, createdUser.ID, 5)
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)
		finalUser, err := testUserStore.GetByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), finalUser.FollowerCount)
	})

	t.Run("データベース切断時、ラップされたエラーを返すこと", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		tempUserStore := NewPostgresUserStore(tempDB)

		tempDB.Close()
		err = tempUserStore.IncreaseFollowerCount(ctx, createdUser.ID, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "follower countsの更新に失敗しました")
		t.Logf("エラーは: %v\n", err)
	})
}
