package entities_test

import (
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// TransferRequest Entity Tests
// ========================================

func TestNewTransferRequest(t *testing.T) {
	t.Run("正常に送金リクエストを作成", func(t *testing.T) {
		fromUserID := uuid.New()
		toUserID := uuid.New()
		amount := int64(1000)
		message := "Test transfer"
		idempotencyKey := "test-key-123"

		tr, err := entities.NewTransferRequest(fromUserID, toUserID, amount, message, idempotencyKey)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, tr.ID)
		assert.Equal(t, fromUserID, tr.FromUserID)
		assert.Equal(t, toUserID, tr.ToUserID)
		assert.Equal(t, amount, tr.Amount)
		assert.Equal(t, message, tr.Message)
		assert.Equal(t, idempotencyKey, tr.IdempotencyKey)
		assert.Equal(t, entities.TransferRequestStatusPending, tr.Status)
		assert.False(t, tr.CreatedAt.IsZero())
		assert.False(t, tr.UpdatedAt.IsZero())
		assert.False(t, tr.ExpiresAt.IsZero())
		assert.Nil(t, tr.ApprovedAt)
		assert.Nil(t, tr.RejectedAt)
		assert.Nil(t, tr.CancelledAt)
		assert.Nil(t, tr.TransactionID)

		// ExpiresAt should be 24 hours from now
		expectedExpiry := time.Now().Add(24 * time.Hour)
		assert.WithinDuration(t, expectedExpiry, tr.ExpiresAt, 5*time.Second)
	})

	t.Run("金額が0以下はエラー", func(t *testing.T) {
		_, err := entities.NewTransferRequest(uuid.New(), uuid.New(), 0, "test", "key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")

		_, err = entities.NewTransferRequest(uuid.New(), uuid.New(), -100, "test", "key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("送信者と受信者が同じはエラー", func(t *testing.T) {
		userID := uuid.New()
		_, err := entities.NewTransferRequest(userID, userID, 1000, "test", "key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot send to yourself")
	})

	t.Run("冪等性キーが空はエラー", func(t *testing.T) {
		_, err := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "idempotency_key is required")
	})
}

func TestTransferRequest_Approve(t *testing.T) {
	t.Run("pending状態のリクエストを承認", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		transactionID := uuid.New()

		err := tr.Approve(transactionID)
		require.NoError(t, err)

		assert.Equal(t, entities.TransferRequestStatusApproved, tr.Status)
		assert.NotNil(t, tr.ApprovedAt)
		assert.NotNil(t, tr.TransactionID)
		assert.Equal(t, transactionID, *tr.TransactionID)
	})

	t.Run("approved状態からの承認はエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		transactionID := uuid.New()
		tr.Approve(transactionID)

		err := tr.Approve(uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})

	t.Run("rejected状態からの承認はエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Reject()

		err := tr.Approve(uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})

	t.Run("cancelled状態からの承認はエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Cancel()

		err := tr.Approve(uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})

	t.Run("期限切れの場合承認はエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour) // 期限切れに設定

		err := tr.Approve(uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request has expired")
	})
}

func TestTransferRequest_Reject(t *testing.T) {
	t.Run("pending状態のリクエストを拒否", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")

		err := tr.Reject()
		require.NoError(t, err)

		assert.Equal(t, entities.TransferRequestStatusRejected, tr.Status)
		assert.NotNil(t, tr.RejectedAt)
	})

	t.Run("approved状態からの拒否はエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())

		err := tr.Reject()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})
}

func TestTransferRequest_Cancel(t *testing.T) {
	t.Run("pending状態のリクエストをキャンセル", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")

		err := tr.Cancel()
		require.NoError(t, err)

		assert.Equal(t, entities.TransferRequestStatusCancelled, tr.Status)
		assert.NotNil(t, tr.CancelledAt)
	})

	t.Run("approved状態からのキャンセルはエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())

		err := tr.Cancel()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})

	t.Run("rejected状態からのキャンセルはエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Reject()

		err := tr.Cancel()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})
}

func TestTransferRequest_MarkAsExpired(t *testing.T) {
	t.Run("リクエストを期限切れにマーク", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour) // 期限切れに設定

		tr.MarkAsExpired()

		assert.Equal(t, entities.TransferRequestStatusExpired, tr.Status)
	})

	t.Run("approved状態のリクエストは期限切れにできない", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())
		originalStatus := tr.Status

		tr.MarkAsExpired()

		assert.Equal(t, originalStatus, tr.Status)
	})
}

func TestTransferRequest_IsExpired(t *testing.T) {
	t.Run("ExpiresAtが過去ならtrue", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour)

		assert.True(t, tr.IsExpired())
	})

	t.Run("ExpiresAtが未来ならfalse", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.ExpiresAt = time.Now().Add(1 * time.Hour)

		assert.False(t, tr.IsExpired())
	})
}

func TestTransferRequest_IsPending(t *testing.T) {
	t.Run("pending状態ならtrue", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")

		assert.True(t, tr.IsPending())
	})

	t.Run("approved状態ならfalse", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())

		assert.False(t, tr.IsPending())
	})

	t.Run("rejected状態ならfalse", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Reject()

		assert.False(t, tr.IsPending())
	})
}

func TestTransferRequest_CanApprove(t *testing.T) {
	t.Run("pending状態で期限内ならエラーなし", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")

		err := tr.CanApprove()
		assert.NoError(t, err)
	})

	t.Run("approved状態ならエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())

		err := tr.CanApprove()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})

	t.Run("期限切れならエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour)

		err := tr.CanApprove()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request has expired")
	})
}

func TestTransferRequest_CanReject(t *testing.T) {
	t.Run("pending状態ならエラーなし", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")

		err := tr.CanReject()
		assert.NoError(t, err)
	})

	t.Run("approved状態ならエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())

		err := tr.CanReject()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})
}

func TestTransferRequest_CanCancel(t *testing.T) {
	t.Run("pending状態ならエラーなし", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")

		err := tr.CanCancel()
		assert.NoError(t, err)
	})

	t.Run("approved状態ならエラー", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(uuid.New(), uuid.New(), 1000, "test", "key-123")
		tr.Approve(uuid.New())

		err := tr.CanCancel()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not pending")
	})
}
