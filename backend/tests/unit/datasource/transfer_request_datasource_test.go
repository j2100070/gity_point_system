//go:build integration
// +build integration

package datasource

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestUserForTransferRequest はTransferRequestテスト用のユーザーを作成（初期残高10000）
func createTestUserForTransferRequest(t *testing.T, db inframysql.DB, username string) *entities.User {
	return createTestUserWithBalanceDB(t, db, username, 10000)
}

// ========================================
// TransferRequestDataSource Insert / Select Tests
// ========================================

func TestTransferRequestDataSource_InsertAndSelect(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender := createTestUserForTransferRequest(t, db, "sender")
	receiver := createTestUserForTransferRequest(t, db, "receiver")

	t.Run("送金リクエストを作成して取得", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test transfer", "key-123")
		err := ds.Insert(ctx, tr)
		require.NoError(t, err)

		retrieved, err := ds.Select(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, tr.ID, retrieved.ID)
		assert.Equal(t, sender.ID, retrieved.FromUserID)
		assert.Equal(t, receiver.ID, retrieved.ToUserID)
		assert.Equal(t, int64(1000), retrieved.Amount)
		assert.Equal(t, "Test transfer", retrieved.Message)
		assert.Equal(t, "key-123", retrieved.IdempotencyKey)
		assert.Equal(t, entities.TransferRequestStatusPending, retrieved.Status)
	})

	t.Run("存在しないIDはnil", func(t *testing.T) {
		result, err := ds.Select(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestTransferRequestDataSource_SelectByIdempotencyKey(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender := createTestUserForTransferRequest(t, db, "sender")
	receiver := createTestUserForTransferRequest(t, db, "receiver")

	t.Run("冪等性キーで検索", func(t *testing.T) {
		idempotencyKey := "unique-key-456"
		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 2000, "Test", idempotencyKey)
		require.NoError(t, ds.Insert(ctx, tr))

		found, err := ds.SelectByIdempotencyKey(ctx, idempotencyKey)
		require.NoError(t, err)
		assert.Equal(t, tr.ID, found.ID)
		assert.Equal(t, idempotencyKey, found.IdempotencyKey)
	})

	t.Run("存在しないキーはnil", func(t *testing.T) {
		found, err := ds.SelectByIdempotencyKey(ctx, "non-existent-key")
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestTransferRequestDataSource_Update(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender := createTestUserForTransferRequest(t, db, "sender")
	receiver := createTestUserForTransferRequest(t, db, "receiver")

	t.Run("ステータスを更新", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-update")
		require.NoError(t, ds.Insert(ctx, tr))

		// Approve
		transactionID := uuid.New()
		tr.Approve(transactionID)
		err := ds.Update(ctx, tr)
		require.NoError(t, err)

		retrieved, _ := ds.Select(ctx, tr.ID)
		assert.Equal(t, entities.TransferRequestStatusApproved, retrieved.Status)
		assert.NotNil(t, retrieved.ApprovedAt)
		assert.NotNil(t, retrieved.TransactionID)
		assert.Equal(t, transactionID, *retrieved.TransactionID)
	})
}

func TestTransferRequestDataSource_SelectPendingByToUser(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender1 := createTestUserForTransferRequest(t, db, "sender1")
	sender2 := createTestUserForTransferRequest(t, db, "sender2")
	receiver := createTestUserForTransferRequest(t, db, "receiver")

	t.Run("受信者の承認待ちリクエストを取得", func(t *testing.T) {
		// 3つのリクエストを作成
		tr1, _ := entities.NewTransferRequest(sender1.ID, receiver.ID, 1000, "Request 1", "key-1")
		tr2, _ := entities.NewTransferRequest(sender2.ID, receiver.ID, 2000, "Request 2", "key-2")
		tr3, _ := entities.NewTransferRequest(sender1.ID, receiver.ID, 3000, "Request 3", "key-3")

		require.NoError(t, ds.Insert(ctx, tr1))
		require.NoError(t, ds.Insert(ctx, tr2))
		require.NoError(t, ds.Insert(ctx, tr3))

		// tr1を承認済みにする
		tr1.Approve(uuid.New())
		ds.Update(ctx, tr1)

		// 承認待ちリクエストを取得
		pending, err := ds.SelectPendingByToUser(ctx, receiver.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, pending, 2) // tr2とtr3のみ

		// 新しい順にソートされているか確認
		if len(pending) == 2 {
			assert.Equal(t, tr3.ID, pending[0].ID)
			assert.Equal(t, tr2.ID, pending[1].ID)
		}
	})

	t.Run("ページネーション", func(t *testing.T) {
		db := setupTransferRequestTestDB(t)
		ctx := context.Background()

		ds := dsmysqlimpl.NewTransferRequestDataSource(db)
		sender := createTestUserForTransferRequest(t, db, "sender_page")
		receiver := createTestUserForTransferRequest(t, db, "receiver_page")

		// 5つのリクエストを作成
		for i := 0; i < 5; i++ {
			tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, int64(1000+i*100), "Request", fmt.Sprintf("key-page-%d", i))
			require.NoError(t, ds.Insert(ctx, tr))
			time.Sleep(10 * time.Millisecond) // 作成時刻を少しずらす
		}

		// 最初の2件
		page1, err := ds.SelectPendingByToUser(ctx, receiver.ID, 0, 2)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		// 次の2件
		page2, err := ds.SelectPendingByToUser(ctx, receiver.ID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		// 最後の1件
		page3, err := ds.SelectPendingByToUser(ctx, receiver.ID, 4, 2)
		require.NoError(t, err)
		assert.Len(t, page3, 1)
	})
}

func TestTransferRequestDataSource_SelectSentByFromUser(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender := createTestUserForTransferRequest(t, db, "sender")
	receiver1 := createTestUserForTransferRequest(t, db, "receiver1")
	receiver2 := createTestUserForTransferRequest(t, db, "receiver2")

	t.Run("送信者の送信リクエストを取得", func(t *testing.T) {
		tr1, _ := entities.NewTransferRequest(sender.ID, receiver1.ID, 1000, "To R1", "key-sent-1")
		tr2, _ := entities.NewTransferRequest(sender.ID, receiver2.ID, 2000, "To R2", "key-sent-2")

		require.NoError(t, ds.Insert(ctx, tr1))
		require.NoError(t, ds.Insert(ctx, tr2))

		sent, err := ds.SelectSentByFromUser(ctx, sender.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, sent, 2)
	})
}

func TestTransferRequestDataSource_CountPendingByToUser(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender := createTestUserForTransferRequest(t, db, "sender_count")
	receiver := createTestUserForTransferRequest(t, db, "receiver_count")

	t.Run("承認待ちリクエスト数をカウント", func(t *testing.T) {
		// 3つのpendingリクエストを作成
		for i := 0; i < 3; i++ {
			tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", fmt.Sprintf("key-count-%d", i))
			require.NoError(t, ds.Insert(ctx, tr))
		}

		// 1つを承認済みにする
		tr4, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-count-4")
		require.NoError(t, ds.Insert(ctx, tr4))
		tr4.Approve(uuid.New())
		ds.Update(ctx, tr4)

		// カウント
		count, err := ds.CountPendingByToUser(ctx, receiver.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("pendingがない場合は0", func(t *testing.T) {
		newReceiver := createTestUserForTransferRequest(t, db, "receiver_empty")

		count, err := ds.CountPendingByToUser(ctx, newReceiver.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestTransferRequestDataSource_UpdateExpiredRequests(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewTransferRequestDataSource(db)
	sender := createTestUserForTransferRequest(t, db, "sender_expire")
	receiver := createTestUserForTransferRequest(t, db, "receiver_expire")

	t.Run("期限切れリクエストを一括更新", func(t *testing.T) {
		// 期限切れのリクエスト
		tr1, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Expired", "key-expired-1")
		tr1.ExpiresAt = time.Now().Add(-1 * time.Hour)
		require.NoError(t, ds.Insert(ctx, tr1))

		tr2, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 2000, "Expired", "key-expired-2")
		tr2.ExpiresAt = time.Now().Add(-2 * time.Hour)
		require.NoError(t, ds.Insert(ctx, tr2))

		// 期限内のリクエスト
		tr3, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 3000, "Valid", "key-valid")
		require.NoError(t, ds.Insert(ctx, tr3))

		// 期限切れを更新
		updatedCount, err := ds.UpdateExpiredRequests(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), updatedCount)

		// 確認
		retrieved1, _ := ds.Select(ctx, tr1.ID)
		assert.Equal(t, entities.TransferRequestStatusExpired, retrieved1.Status)

		retrieved2, _ := ds.Select(ctx, tr2.ID)
		assert.Equal(t, entities.TransferRequestStatusExpired, retrieved2.Status)

		retrieved3, _ := ds.Select(ctx, tr3.ID)
		assert.Equal(t, entities.TransferRequestStatusPending, retrieved3.Status)
	})

	t.Run("承認済みリクエストは期限切れにしない", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Approved", "key-approved-expired")
		require.NoError(t, ds.Insert(ctx, tr))

		// 承認
		tr.Approve(uuid.New())
		ds.Update(ctx, tr)

		// ExpiresAtを過去に設定
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour)
		ds.Update(ctx, tr)

		// 期限切れ更新を試みる
		ds.UpdateExpiredRequests(ctx)

		// まだ承認済みのまま
		retrieved, _ := ds.Select(ctx, tr.ID)
		assert.Equal(t, entities.TransferRequestStatusApproved, retrieved.Status)
	})
}

// ========================================
// TransferRequest Integration with User
// ========================================

func TestTransferRequestIntegration_WithUsers(t *testing.T) {
	db := setupTransferRequestTestDB(t)
	ctx := context.Background()

	trDS := dsmysqlimpl.NewTransferRequestDataSource(db)
	userDS := dsmysqlimpl.NewUserDataSource(db)

	sender := createTestUserForTransferRequest(t, db, "integration_sender")
	receiver := createTestUserForTransferRequest(t, db, "integration_receiver")

	t.Run("リクエスト作成からユーザー情報取得まで", func(t *testing.T) {
		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 5000, "Integration test", "key-integration")
		require.NoError(t, trDS.Insert(ctx, tr))

		// リクエスト取得
		retrieved, err := trDS.Select(ctx, tr.ID)
		require.NoError(t, err)

		// ユーザー情報取得
		senderUser, err := userDS.Select(ctx, retrieved.FromUserID)
		require.NoError(t, err)
		assert.Equal(t, sender.Username, senderUser.Username)

		receiverUser, err := userDS.Select(ctx, retrieved.ToUserID)
		require.NoError(t, err)
		assert.Equal(t, receiver.Username, receiverUser.Username)
	})
}
