package db
import(
	"aita/internal/models"
	"context"
	"testing"
	"github.com/stretchr/testify/require"
)
func TestCreateUser(t *testing.T) {
	defer cleanUpTestDB()
	req := &models.SignupRequest{
		Username: "henry",
		Email: "text@example.com",
		Password:"a123456",
	}

	createdUser, err := testStore.Create(context.Background(), req)
	require.NoError(t, err, "CreateUserはエラーを返すべきではありません")
	require.NotNil(t, createdUser,"作成されたユーザーオブジェクトは空であるべきではありません")
	require.Equal(t, req.Username, createdUser.Username,"ユーザー名は一致する必要があります")
	require.Equal(t, req.Email, createdUser.Email,"メールは一致する必要があります")
	require.NotEmpty(t, createdUser.PasswordHash,"パスワードハッシュは空であるべきではありません")
	require.NotEqual(t, req.Password, createdUser.PasswordHash,"パスワードは決して平文で保存してはいけません")
}

func TestGetByEmail(t *testing.T) {
	defer cleanUpTestDB()
	req := &models.SignupRequest{
		Username: "testuser_for_get",
		Email:    "get@example.com",
		Password: "password123",
	}
	createdUser, err := testStore.Create(context.Background(),req)
	require.NoError(t,err)
	t.Run("User Found", func(t *testing.T) {
		foundUser, err := testStore.GetByEmail(context.Background(), "get@example.com")
		require.NoError(t, err,"ユーザーが存在する場合、エラーを返すべきではありません")
		require.NotNil(t,foundUser,"見つかったユーザーは空であってはなりません")
		require.Equal(t,createdUser.ID, foundUser.ID,"見つかったユーザーIDが一致するはずです")
		require.Equal(t,createdUser.Username, foundUser.Username,"見つかったユーザー名が一致するはずです")
	})
	t.Run("User Not Found", func(t *testing.T) {
		foundUser, err := testStore.GetByEmail(context.Background(), "nonexistent@example.com")
		require.Error(t,err,"ユーザーが存在しない場合、エラーを返すべきです")
		require.Equal(t,"ユーザーが存在しません",err.Error(),"エラーメッセージは「ユーザーが存在しません」であるべきです")
		require.Nil(t,foundUser,"見つかったユーザーオブジェクトは空であるべきです")
	})
}
