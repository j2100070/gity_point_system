package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// User Entity Tests (Transaction Focus)
// ========================================

func TestUser_CanTransfer(t *testing.T) {
	user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
	user.Balance = 5000

	t.Run("十分な残高がある場合は送金可能", func(t *testing.T) {
		err := user.CanTransfer(3000)
		assert.NoError(t, err)
	})

	t.Run("残高と同じ金額も送金可能", func(t *testing.T) {
		err := user.CanTransfer(5000)
		assert.NoError(t, err)
	})

	t.Run("残高不足の場合はエラー", func(t *testing.T) {
		err := user.CanTransfer(5001)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("0以下の金額はエラー", func(t *testing.T) {
		err := user.CanTransfer(0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")

		err = user.CanTransfer(-100)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("無効なユーザーは送金不可", func(t *testing.T) {
		user.Deactivate()
		err := user.CanTransfer(1000)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not active")
	})
}

func TestUser_Deduct(t *testing.T) {
	t.Run("正常なポイント減算", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 5000
		initialVersion := user.Version

		err := user.Deduct(2000)

		require.NoError(t, err)
		assert.Equal(t, int64(3000), user.Balance)
		assert.Equal(t, initialVersion+1, user.Version) // バージョン増加
	})

	t.Run("残高不足で減算失敗", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 1000
		initialBalance := user.Balance
		initialVersion := user.Version

		err := user.Deduct(1500)

		require.Error(t, err)
		assert.Equal(t, initialBalance, user.Balance) // 残高は変わらない
		assert.Equal(t, initialVersion, user.Version) // バージョンも変わらない
	})

	t.Run("全額減算", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 1000

		err := user.Deduct(1000)

		require.NoError(t, err)
		assert.Equal(t, int64(0), user.Balance)
	})

	t.Run("無効なユーザーは減算不可", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 5000
		user.Deactivate()

		err := user.Deduct(1000)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not active")
	})
}

func TestUser_Add(t *testing.T) {
	t.Run("正常なポイント加算", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 1000
		initialVersion := user.Version

		err := user.Add(500)

		require.NoError(t, err)
		assert.Equal(t, int64(1500), user.Balance)
		assert.Equal(t, initialVersion+1, user.Version)
	})

	t.Run("0からの加算", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 0

		err := user.Add(1000)

		require.NoError(t, err)
		assert.Equal(t, int64(1000), user.Balance)
	})

	t.Run("0以下の金額は加算不可", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		initialBalance := user.Balance

		err := user.Add(0)
		require.Error(t, err)
		assert.Equal(t, initialBalance, user.Balance)

		err = user.Add(-100)
		require.Error(t, err)
		assert.Equal(t, initialBalance, user.Balance)
	})
}

func TestUser_TransactionSequence(t *testing.T) {
	t.Run("連続したトランザクション操作のテスト", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 10000
		initialVersion := user.Version

		// 1. 2000ポイント減算
		err := user.Deduct(2000)
		require.NoError(t, err)
		assert.Equal(t, int64(8000), user.Balance)
		assert.Equal(t, initialVersion+1, user.Version)

		// 2. 1500ポイント加算
		err = user.Add(1500)
		require.NoError(t, err)
		assert.Equal(t, int64(9500), user.Balance)
		assert.Equal(t, initialVersion+2, user.Version)

		// 3. 5000ポイント減算
		err = user.Deduct(5000)
		require.NoError(t, err)
		assert.Equal(t, int64(4500), user.Balance)
		assert.Equal(t, initialVersion+3, user.Version)

		// 4. 残高不足で減算失敗
		err = user.Deduct(5000)
		require.Error(t, err)
		assert.Equal(t, int64(4500), user.Balance) // 残高変わらず
		assert.Equal(t, initialVersion+3, user.Version) // バージョンも変わらず
	})
}

func TestUser_Version_OptimisticLocking(t *testing.T) {
	t.Run("バージョン番号の正しい増加", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 10000
		initialVersion := user.Version

		// 複数の操作でバージョンが増加
		user.Deduct(1000)
		assert.Equal(t, initialVersion+1, user.Version)

		user.Add(500)
		assert.Equal(t, initialVersion+2, user.Version)

		user.Deduct(2000)
		assert.Equal(t, initialVersion+3, user.Version)
	})

	t.Run("失敗した操作ではバージョンが増加しない", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 1000
		initialVersion := user.Version

		// 残高不足で失敗
		err := user.Deduct(2000)
		require.Error(t, err)
		assert.Equal(t, initialVersion, user.Version) // バージョン変わらず

		// 無効な金額で失敗
		err = user.Add(-100)
		require.Error(t, err)
		assert.Equal(t, initialVersion, user.Version) // バージョン変わらず
	})
}

func TestUser_ConcurrentTransactionScenario(t *testing.T) {
	t.Run("同時トランザクションのシミュレーション", func(t *testing.T) {
		// ユーザーAとユーザーBのトランザクションシナリオ
		userA, _ := NewUser("userA", "a@example.com", "hash1", "User A")
		userB, _ := NewUser("userB", "b@example.com", "hash2", "User B")
		userA.Balance = 5000
		userB.Balance = 3000

		// userAがuserBに2000ポイント送る
		err := userA.Deduct(2000)
		require.NoError(t, err)

		err = userB.Add(2000)
		require.NoError(t, err)

		// 結果確認
		assert.Equal(t, int64(3000), userA.Balance)
		assert.Equal(t, int64(5000), userB.Balance)

		// ポイントの合計は保存される（8000ポイント）
		total := userA.Balance + userB.Balance
		assert.Equal(t, int64(8000), total)
	})
}

func TestUser_EdgeCases(t *testing.T) {
	t.Run("最小金額での操作", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 1

		err := user.Deduct(1)
		require.NoError(t, err)
		assert.Equal(t, int64(0), user.Balance)
	})

	t.Run("大きな金額での操作", func(t *testing.T) {
		user, _ := NewUser("testuser", "test@example.com", "hashedpass", "Test User")
		user.Balance = 9999999999

		err := user.Deduct(5000000000)
		require.NoError(t, err)
		assert.Equal(t, int64(4999999999), user.Balance)
	})
}

func TestNewUser(t *testing.T) {
	t.Run("正常なユーザー作成", func(t *testing.T) {
		user, err := NewUser("testuser", "test@example.com", "hashedpass", "Test User")

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, int64(0), user.Balance) // 初期残高は0
		assert.Equal(t, RoleUser, user.Role)
		assert.Equal(t, 1, user.Version) // 初期バージョンは1
		assert.True(t, user.IsActive)
	})

	t.Run("必須フィールドの検証", func(t *testing.T) {
		_, err := NewUser("", "test@example.com", "hashedpass", "Test User")
		require.Error(t, err)

		_, err = NewUser("testuser", "", "hashedpass", "Test User")
		require.Error(t, err)

		_, err = NewUser("testuser", "test@example.com", "", "Test User")
		require.Error(t, err)

		_, err = NewUser("testuser", "test@example.com", "hashedpass", "")
		require.Error(t, err)
	})
}
