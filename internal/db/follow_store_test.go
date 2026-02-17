package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/testutils"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUser(t *testing.T, ctx context.Context) (userA, userB *models.User){

	u1 := &models.User{
		Username: "userA",
		Email: "a@example.com",
		PasswordHash: "passwordhash1",
	}

	userA, err := testUserStore.Create(ctx, u1)
	require.NoError(t, err)

	u2 := &models.User{
		Username: "userB",
		Email: "b@example.com",
		PasswordHash: "passwordhash2",
	}
	userB, err = testUserStore.Create(ctx, u2)
	require.NoError(t, err)

	return userA, userB
}
func TestCreateFollow(t *testing.T) {
	testContext.CleanupTestDB()
	ctx := context.Background()
	userA, userB := setupUser(t, ctx)
	defer testContext.CleanupTestDB()

	t.Run("正常系: 正常にフォローできること", func(t *testing.T) {
		follow := &models.Follow{
			FollowerID: userA.ID,
			FollowingID: userB.ID,
		}

		createdFollow, err := testFollowStore.Create(ctx, follow)
		require.NoError(t, err)
		require.NotNil(t, createdFollow)
		assert.NotZero(t, createdFollow.ID)
		assert.Equal(t, follow.FollowerID, createdFollow.FollowerID)
		assert.Equal(t, follow.FollowingID, createdFollow.FollowingID)
		assert.NotZero(t, createdFollow.CreatedAt)
		assert.Equal(t, time.UTC, createdFollow.CreatedAt.Location())
	})
}

func TestCreateFollowWhileErr(t *testing.T) {
	testContext.CleanupTestDB()
	ctx := context.Background()
	userA, userB := setupUser(t, ctx)
	defer testContext.CleanupTestDB()

	t.Run("異常系: 重複フォロー", func(t *testing.T) {
		follow := &models.Follow{
			FollowerID: userA.ID,
			FollowingID: userB.ID,
		}
		_, err := testFollowStore.Create(ctx, follow)
		require.NoError(t, err)
		newFollow, err := testFollowStore.Create(ctx, follow)
		assert.ErrorIs(t, err, errcode.ErrAlreadyFollowing)
		assert.Nil(t, newFollow)
	})

	t.Run("異常系: 自分自身をフォローした", func(t *testing.T) {
		abnormaFollow := &models.Follow{
			FollowerID: userA.ID,
			FollowingID: userA.ID,
		}

		follow, err := testFollowStore.Create(ctx, abnormaFollow)
		assert.ErrorIs(t, err, errcode.ErrCannotFollowSelf)
		assert.Nil(t, follow)
	})


	t.Run("異常系: 存在しないユーザー", func(t *testing.T) {
		follow := &models.Follow{
			FollowerID: 999,
			FollowingID: userB.ID,
		}

		newFollow, err := testFollowStore.Create(ctx, follow)
		assert.ErrorIs(t, err, errcode.ErrUserNotFound)
		assert.Nil(t, newFollow)
	})

	t.Run("異常系: データベース内部エラー", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)

		tempFollowStore := NewPostgresFollowStore(tempDB)
		tempDB.Close()

		follow := &models.Follow{
			FollowerID: userB.ID,
			FollowingID: userA.ID,
		}

		createdFollow, err := tempFollowStore.Create(ctx, follow)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "フォローの生成に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, createdFollow)

	})
}

func TestGetFollowing(t *testing.T) {
	testContext.CleanupTestDB()
	ctx := context.Background()
	userA, userB := setupUser(t, ctx)
	defer testContext.CleanupTestDB()

	u3 := &models.User{
		Username: "userC",
		Email: "c@example.com",
		PasswordHash: "passwordhash3",
	}

	userC, err := testUserStore.Create(ctx, u3)
	require.NoError(t, err)

	followAB := &models.Follow{
			FollowerID: userA.ID,
			FollowingID: userB.ID,
		}
	_, err = testFollowStore.Create(ctx, followAB) 
	require.NoError(t, err)
	followAC := &models.Follow{
		FollowerID: userA.ID,
		FollowingID: userC.ID,
	}
	_, err = testFollowStore.Create(ctx, followAC)
	require.NoError(t, err)

	t.Run("正常系：フォロー中リストの取得に成功", func(t *testing.T) {
		res, err := testFollowStore.GetFollowing(ctx, userA.ID)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		targetIDs := []int64{res[0].FollowingID, res[1].FollowingID}
		assert.Contains(t, targetIDs, userB.ID)
		assert.Contains(t, targetIDs, userC.ID)
		assert.Equal(t, userA.ID, res[0].FollowerID)
		assert.Equal(t, userA.ID, res[1].FollowerID)
	})

	t.Run("正常系：フォローしているユーザーがいない場合は空のスライスを返すこと", func(t *testing.T) {
        followings, err := testFollowStore.GetFollowing(ctx, userC.ID)
        
        require.NoError(t, err)
    	require.NotNil(t, followings, "nilではなく空のスライスであるべきです")
        assert.Len(t, followings, 0)
    })

	t.Run("異常系: サーバー内部エラー", func(t *testing.T) {
		tempDB, err := testutils.OpenDB(testContext.DSN)
		require.NoError(t, err)
		tempFollowStore := NewPostgresFollowStore(tempDB)
		tempDB.Close()

		followings, err := tempFollowStore.GetFollowing(ctx, userC.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "フォロー中リストの取得に失敗しました")
		t.Logf("エラーは: %v\n", err)
		assert.Nil(t, followings)

	})

}

func TestGetFollowers(t *testing.T) {
	testContext.CleanupTestDB()
	ctx := context.Background()
	userA, userB := setupUser(t, ctx)
	defer testContext.CleanupTestDB()

	u3 := &models.User{
		Username: "userC",
		Email: "c@example.com",
		PasswordHash: "passwordhash3",
	}

	userC, err := testUserStore.Create(ctx, u3)
	require.NoError(t, err)

	followBA := &models.Follow{
			FollowerID: userB.ID,
			FollowingID: userA.ID,
		}
	_, err = testFollowStore.Create(ctx, followBA) 
	require.NoError(t, err)
	followCA := &models.Follow{
		FollowerID: userC.ID,
		FollowingID: userA.ID,
	}
	_, err = testFollowStore.Create(ctx, followCA)
	require.NoError(t, err)

	t.Run("正常系: ", func(t *testing.T) {
		followers, err := testFollowStore.GetFollower(ctx, userA.ID)
		require.NoError(t, err)
		require.NotNil(t, followers)
		assert.Len(t, followers, 2)
		targetIDs := []int64{followers[0].FollowerID, followers[1].FollowerID}
		actualIDs := []int64{userC.ID, userB.ID}
		assert.ElementsMatch(t, actualIDs, targetIDs)
		assert.Equal(t, userA.ID, followers[0].FollowingID)
		assert.Equal(t, userA.ID, followers[0].FollowingID)
	})

	t.Run("正常系:フォローされているユーザーがいない場合は空のスライスを返すこと", func(t *testing.T) {
		followers, err := testFollowStore.GetFollower(ctx, userC.ID)
        
        require.NoError(t, err)
    	require.NotNil(t, followers, "nilではなく空のスライスであるべきです")
        assert.Len(t, followers, 0)
	})

    t.Run("異常系: サーバー内部エラー(DB接続終了後)", func(t *testing.T) {
        tempDB, err := testutils.OpenDB(testContext.DSN)
        require.NoError(t, err)
        tempFollowStore := NewPostgresFollowStore(tempDB)
        tempDB.Close() 

        followers, err := tempFollowStore.GetFollower(ctx, userA.ID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "フォロワーリストの取得に失敗しました")
        assert.Nil(t, followers)
        t.Logf("期待通りのエラーを確認: %v\n", err)
    })
}

func TestGetRelationship(t *testing.T) {
    testContext.CleanupTestDB()
    ctx := context.Background()
    userA, userB := setupUser(t, ctx)
    defer testContext.CleanupTestDB()

    t.Run("正常系 :何の関わりもない場合 (false, false)", func(t *testing.T) {
        testContext.CleanupTestDB()
        rel, err := testFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.NoError(t, err)
        require.NotNil(t, rel)
        assert.False(t, rel.Following, "Followingはfalseであるべきです")
        assert.False(t, rel.FollowedBy, "FollowedByはfalseであるべきです")
    })

    t.Run("正常系: AがBをフォローしている場合 (true, false)", func(t *testing.T) {
		userA, userB := setupUser(t, ctx)
		_, err := testFollowStore.Create(ctx, &models.Follow{
            FollowerID:  userA.ID,
            FollowingID: userB.ID,
        })
        require.NoError(t, err)

        rel, err := testFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.NoError(t, err)
        assert.True(t, rel.Following, "AがBをフォローしているのでtrue")
        assert.False(t, rel.FollowedBy, "BはAをフォローしていないのでfalse")
    })

    t.Run("正常系: BがAをフォローしている場合 (false, true)", func(t *testing.T) {
        testContext.CleanupTestDB()
        userA, userB = setupUser(t, ctx)

        _, err := testFollowStore.Create(ctx, &models.Follow{
            FollowerID:  userB.ID,
            FollowingID: userA.ID,
        })
        require.NoError(t, err)

        rel, err := testFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.NoError(t, err)
        assert.False(t, rel.Following, "AはBをフォローしていないのでfalse")
        assert.True(t, rel.FollowedBy, "BがAをフォローしているのでtrue")
    })

    t.Run("正常系: 互いにフォローしている場合 (true, true)", func(t *testing.T) {
        _, err := testFollowStore.Create(ctx, &models.Follow{
            FollowerID:  userA.ID,
            FollowingID: userB.ID,
        })
        require.NoError(t, err)

        rel, err := testFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.NoError(t, err)
        assert.True(t, rel.Following, "相互フォローなのでtrue")
        assert.True(t, rel.FollowedBy, "相互フォローなのでtrue")
    })

    t.Run("異常系：データベース接続エラー", func(t *testing.T) {
        tempDB, _ := testutils.OpenDB(testContext.DSN)
        tempFollowStore := NewPostgresFollowStore(tempDB)
        tempDB.Close() 

        rel, err := tempFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "関係性の取得に失敗しました")
        assert.Nil(t, rel)
    })
}



func TestDelete(t *testing.T) {
    testContext.CleanupTestDB()
    ctx := context.Background()
    userA, userB := setupUser(t, ctx)
    defer testContext.CleanupTestDB()

    _, err := testFollowStore.Create(ctx, &models.Follow{
        FollowerID:  userA.ID,
        FollowingID: userB.ID,
    })
    require.NoError(t, err)

    t.Run("正常系: フォロー解除が成功し、関係性が更新されること", func(t *testing.T) {
        relBefore, err := testFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.NoError(t, err)
        assert.True(t, relBefore.Following)

        err = testFollowStore.Delete(ctx, userA.ID, userB.ID)
        require.NoError(t, err)

        relAfter, err := testFollowStore.GetRelationship(ctx, userA.ID, userB.ID)
        require.NoError(t, err)
        assert.False(t, relAfter.Following, "削除後はFollowingがfalseになるべきです")
    })

    t.Run("正常系: 存在しないフォロー関係を解除してもエラーにならないこと", func(t *testing.T) {
        err := testFollowStore.Delete(ctx, userB.ID, userA.ID)
        require.NoError(t, err, "存在しない関係の削除も成功すべきです")
    })

    t.Run("異常系: サーバー内部エラー (DB切断)", func(t *testing.T) {
        tempDB, _ := testutils.OpenDB(testContext.DSN)
        tempFollowStore := NewPostgresFollowStore(tempDB)
        tempDB.Close()

        err := tempFollowStore.Delete(ctx, userA.ID, userB.ID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "フォロー解除に失敗しました")
    })
}