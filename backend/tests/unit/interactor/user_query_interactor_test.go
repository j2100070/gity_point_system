package interactor_test

import (
	"context"
	"testing"

	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// UserQueryInteractor テスト
// ========================================

// --- GetUserByID ---

func TestUserQueryInteractor_GetUserByID(t *testing.T) {
	t.Run("正常にユーザー情報を取得できる", func(t *testing.T) {
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewUserQueryInteractor(userRepo, &mockLogger{})

		user := createTestUserWithBalance(t, "queryuser", 5000, "user")
		userRepo.setUser(user)

		resp, err := sut.GetUserByID(context.Background(), &inputport.GetUserByIDRequest{
			UserID: user.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, user.ID, resp.User.ID)
		assert.Equal(t, "queryuser", resp.User.Username)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		sut := interactor.NewUserQueryInteractor(newCtxTrackingUserRepo(), &mockLogger{})

		_, err := sut.GetUserByID(context.Background(), &inputport.GetUserByIDRequest{
			UserID: uuid.New(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

// --- SearchUserByUsername ---

func TestUserQueryInteractor_SearchUserByUsername(t *testing.T) {
	t.Run("正常にユーザーを検索できる", func(t *testing.T) {
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewUserQueryInteractor(userRepo, &mockLogger{})

		user := createTestUserWithBalance(t, "searchable", 1000, "user")
		userRepo.setUser(user)

		resp, err := sut.SearchUserByUsername(context.Background(), &inputport.SearchUserByUsernameRequest{
			Username: "searchable",
		})
		require.NoError(t, err)
		assert.Equal(t, user.ID, resp.User.ID)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		sut := interactor.NewUserQueryInteractor(newCtxTrackingUserRepo(), &mockLogger{})

		_, err := sut.SearchUserByUsername(context.Background(), &inputport.SearchUserByUsernameRequest{
			Username: "nonexistent",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}
