package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Transaction Entity Tests
// ========================================

func TestNewTransfer(t *testing.T) {
	fromUserID := uuid.New()
	toUserID := uuid.New()
	amount := int64(1000)
	idempotencyKey := "test-key-123"
	description := "Test transfer"

	t.Run("正常なトランザクション作成", func(t *testing.T) {
		tx, err := NewTransfer(fromUserID, toUserID, amount, idempotencyKey, description)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tx.ID)
		assert.Equal(t, fromUserID, *tx.FromUserID)
		assert.Equal(t, toUserID, *tx.ToUserID)
		assert.Equal(t, amount, tx.Amount)
		assert.Equal(t, TransactionTypeTransfer, tx.TransactionType)
		assert.Equal(t, TransactionStatusPending, tx.Status)
		assert.Equal(t, idempotencyKey, *tx.IdempotencyKey)
		assert.Equal(t, description, tx.Description)
		assert.NotNil(t, tx.Metadata)
		assert.Nil(t, tx.CompletedAt)
	})

	t.Run("同じユーザーへの送金はエラー", func(t *testing.T) {
		sameUserID := uuid.New()
		_, err := NewTransfer(sameUserID, sameUserID, amount, idempotencyKey, description)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot transfer to the same user")
	})

	t.Run("0以下の金額はエラー", func(t *testing.T) {
		testCases := []int64{0, -100, -1}

		for _, amt := range testCases {
			_, err := NewTransfer(fromUserID, toUserID, amt, idempotencyKey, description)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "amount must be positive")
		}
	})

	t.Run("空の冪等性キーはエラー", func(t *testing.T) {
		_, err := NewTransfer(fromUserID, toUserID, amount, "", description)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "idempotency key is required")
	})
}

func TestNewAdminGrant(t *testing.T) {
	toUserID := uuid.New()
	adminID := uuid.New()
	amount := int64(5000)
	description := "Admin grant points"

	t.Run("正常な管理者付与トランザクション作成", func(t *testing.T) {
		tx, err := NewAdminGrant(toUserID, amount, description, adminID)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tx.ID)
		assert.Nil(t, tx.FromUserID) // システムからの付与
		assert.Equal(t, toUserID, *tx.ToUserID)
		assert.Equal(t, amount, tx.Amount)
		assert.Equal(t, TransactionTypeAdminGrant, tx.TransactionType)
		assert.Equal(t, TransactionStatusCompleted, tx.Status) // すぐに完了状態
		assert.NotNil(t, tx.CompletedAt)
		assert.Equal(t, adminID.String(), tx.Metadata["admin_id"])
	})

	t.Run("0以下の金額はエラー", func(t *testing.T) {
		_, err := NewAdminGrant(toUserID, 0, description, adminID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")

		_, err = NewAdminGrant(toUserID, -100, description, adminID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

func TestNewAdminDeduct(t *testing.T) {
	fromUserID := uuid.New()
	adminID := uuid.New()
	amount := int64(2000)
	description := "Admin deduct points"

	t.Run("正常な管理者減算トランザクション作成", func(t *testing.T) {
		tx, err := NewAdminDeduct(fromUserID, amount, description, adminID)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tx.ID)
		assert.Equal(t, fromUserID, *tx.FromUserID)
		assert.Nil(t, tx.ToUserID) // システムへの返却
		assert.Equal(t, amount, tx.Amount)
		assert.Equal(t, TransactionTypeAdminDeduct, tx.TransactionType)
		assert.Equal(t, TransactionStatusCompleted, tx.Status)
		assert.NotNil(t, tx.CompletedAt)
		assert.Equal(t, adminID.String(), tx.Metadata["admin_id"])
	})

	t.Run("0以下の金額はエラー", func(t *testing.T) {
		_, err := NewAdminDeduct(fromUserID, 0, description, adminID)
		require.Error(t, err)

		_, err = NewAdminDeduct(fromUserID, -500, description, adminID)
		require.Error(t, err)
	})
}

func TestTransaction_Complete(t *testing.T) {
	fromUserID := uuid.New()
	toUserID := uuid.New()

	t.Run("Pending状態のトランザクションを完了できる", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-1", "test")
		assert.Equal(t, TransactionStatusPending, tx.Status)
		assert.Nil(t, tx.CompletedAt)

		beforeComplete := time.Now()
		err := tx.Complete()

		require.NoError(t, err)
		assert.Equal(t, TransactionStatusCompleted, tx.Status)
		assert.NotNil(t, tx.CompletedAt)
		assert.True(t, tx.CompletedAt.After(beforeComplete) || tx.CompletedAt.Equal(beforeComplete))
	})

	t.Run("Completed状態のトランザクションは再度完了できない", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-2", "test")
		tx.Complete()

		err := tx.Complete()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in pending status")
	})

	t.Run("Failed状態のトランザクションは完了できない", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-3", "test")
		tx.Fail()

		err := tx.Complete()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in pending status")
	})
}

func TestTransaction_Fail(t *testing.T) {
	fromUserID := uuid.New()
	toUserID := uuid.New()

	t.Run("Pending状態のトランザクションを失敗にできる", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-4", "test")
		assert.Equal(t, TransactionStatusPending, tx.Status)

		err := tx.Fail()

		require.NoError(t, err)
		assert.Equal(t, TransactionStatusFailed, tx.Status)
	})

	t.Run("Completed状態のトランザクションは失敗にできない", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-5", "test")
		tx.Complete()

		err := tx.Fail()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in pending status")
	})
}

func TestTransaction_StateTransitions(t *testing.T) {
	fromUserID := uuid.New()
	toUserID := uuid.New()

	t.Run("状態遷移のテスト: Pending -> Completed", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-6", "test")

		assert.Equal(t, TransactionStatusPending, tx.Status)

		err := tx.Complete()
		require.NoError(t, err)
		assert.Equal(t, TransactionStatusCompleted, tx.Status)
	})

	t.Run("状態遷移のテスト: Pending -> Failed", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-7", "test")

		assert.Equal(t, TransactionStatusPending, tx.Status)

		err := tx.Fail()
		require.NoError(t, err)
		assert.Equal(t, TransactionStatusFailed, tx.Status)
	})

	t.Run("不正な状態遷移: Completed -> Failed", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-8", "test")
		tx.Complete()

		err := tx.Fail()
		require.Error(t, err)
	})

	t.Run("不正な状態遷移: Failed -> Completed", func(t *testing.T) {
		tx, _ := NewTransfer(fromUserID, toUserID, 1000, "key-9", "test")
		tx.Fail()

		err := tx.Complete()
		require.Error(t, err)
	})
}

// ========================================
// IdempotencyKey Entity Tests
// ========================================

func TestNewIdempotencyKey(t *testing.T) {
	key := "test-idempotency-key"
	userID := uuid.New()

	t.Run("正常な冪等性キー作成", func(t *testing.T) {
		beforeCreate := time.Now()
		idempKey := NewIdempotencyKey(key, userID)
		afterCreate := time.Now()

		assert.Equal(t, key, idempKey.Key)
		assert.Equal(t, userID, idempKey.UserID)
		assert.Equal(t, "processing", idempKey.Status)
		assert.Nil(t, idempKey.TransactionID)

		assert.True(t, idempKey.CreatedAt.After(beforeCreate) || idempKey.CreatedAt.Equal(beforeCreate))
		assert.True(t, idempKey.CreatedAt.Before(afterCreate) || idempKey.CreatedAt.Equal(afterCreate))

		// 有効期限は24時間後
		expectedExpiry := idempKey.CreatedAt.Add(24 * time.Hour)
		assert.True(t, idempKey.ExpiresAt.Equal(expectedExpiry) ||
			idempKey.ExpiresAt.Sub(expectedExpiry).Abs() < time.Second)
	})
}

func TestIdempotencyKey_ExpiryTime(t *testing.T) {
	key := "test-key"
	userID := uuid.New()

	idempKey := NewIdempotencyKey(key, userID)

	// 有効期限が作成時刻の24時間後であることを確認
	duration := idempKey.ExpiresAt.Sub(idempKey.CreatedAt)
	assert.Equal(t, 24*time.Hour, duration)
}

// ========================================
// Transaction Type Tests
// ========================================

func TestTransactionType_Constants(t *testing.T) {
	t.Run("トランザクションタイプの定数値確認", func(t *testing.T) {
		assert.Equal(t, TransactionType("transfer"), TransactionTypeTransfer)
		assert.Equal(t, TransactionType("admin_grant"), TransactionTypeAdminGrant)
		assert.Equal(t, TransactionType("admin_deduct"), TransactionTypeAdminDeduct)
		assert.Equal(t, TransactionType("system_grant"), TransactionTypeSystemGrant)
	})
}

func TestTransactionStatus_Constants(t *testing.T) {
	t.Run("トランザクション状態の定数値確認", func(t *testing.T) {
		assert.Equal(t, TransactionStatus("pending"), TransactionStatusPending)
		assert.Equal(t, TransactionStatus("completed"), TransactionStatusCompleted)
		assert.Equal(t, TransactionStatus("failed"), TransactionStatusFailed)
		assert.Equal(t, TransactionStatus("reversed"), TransactionStatusReversed)
	})
}

// ========================================
// Edge Cases and Boundary Tests
// ========================================

func TestTransaction_EdgeCases(t *testing.T) {
	fromUserID := uuid.New()
	toUserID := uuid.New()

	t.Run("最小金額でのトランザクション作成", func(t *testing.T) {
		tx, err := NewTransfer(fromUserID, toUserID, 1, "key-min", "Minimum amount")
		require.NoError(t, err)
		assert.Equal(t, int64(1), tx.Amount)
	})

	t.Run("大きな金額でのトランザクション作成", func(t *testing.T) {
		largeAmount := int64(9999999999)
		tx, err := NewTransfer(fromUserID, toUserID, largeAmount, "key-max", "Large amount")
		require.NoError(t, err)
		assert.Equal(t, largeAmount, tx.Amount)
	})

	t.Run("空の説明でもトランザクション作成可能", func(t *testing.T) {
		tx, err := NewTransfer(fromUserID, toUserID, 1000, "key-empty-desc", "")
		require.NoError(t, err)
		assert.Equal(t, "", tx.Description)
	})
}
