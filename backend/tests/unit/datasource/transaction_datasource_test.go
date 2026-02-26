//go:build integration
// +build integration

package datasource

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// TransactionDataSource Insert / Select Tests
// ========================================

func TestTransactionDataSource_InsertAndSelect(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewTransactionDataSource(db)
	sender := createTestUser(t, db, "tx_sender")
	receiver := createTestUser(t, db, "tx_receiver")

	t.Run("トランザクションを作成してIDで取得", func(t *testing.T) {
		tx, err := entities.NewTransfer(sender.ID, receiver.ID, 1000, "key-1", "Integration test transfer")
		require.NoError(t, err)
		tx.Complete()

		err = ds.Insert(context.Background(), tx)
		require.NoError(t, err)

		retrieved, err := ds.Select(context.Background(), tx.ID)
		require.NoError(t, err)
		assert.Equal(t, tx.ID, retrieved.ID)
		assert.Equal(t, int64(1000), retrieved.Amount)
		assert.Equal(t, entities.TransactionTypeTransfer, retrieved.TransactionType)
		assert.Equal(t, entities.TransactionStatusCompleted, retrieved.Status)
		assert.Equal(t, "Integration test transfer", retrieved.Description)
	})

	t.Run("存在しないIDはエラー", func(t *testing.T) {
		_, err := ds.Select(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}

func TestTransactionDataSource_SelectByIdempotencyKey(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewTransactionDataSource(db)
	sender := createTestUser(t, db, "idem_sender")
	receiver := createTestUser(t, db, "idem_receiver")

	t.Run("冪等性キーでトランザクションを検索", func(t *testing.T) {
		key := fmt.Sprintf("idempotency-%d", time.Now().UnixNano())
		tx, _ := entities.NewTransfer(sender.ID, receiver.ID, 500, key, "Idempotency test")
		tx.Complete()
		require.NoError(t, ds.Insert(context.Background(), tx))

		retrieved, err := ds.SelectByIdempotencyKey(context.Background(), key)
		require.NoError(t, err)
		assert.Equal(t, tx.ID, retrieved.ID)
		assert.Equal(t, int64(500), retrieved.Amount)
	})

	t.Run("存在しないキーはエラー", func(t *testing.T) {
		_, err := ds.SelectByIdempotencyKey(context.Background(), "nonexistent-key")
		assert.Error(t, err)
	})
}

// ========================================
// TransactionDataSource List Tests
// ========================================

func TestTransactionDataSource_SelectListByUserID(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewTransactionDataSource(db)
	sender := createTestUser(t, db, "list_sender")
	receiver := createTestUser(t, db, "list_receiver")

	t.Run("ユーザーのトランザクション一覧を取得", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("list-key-%d-%d", i, time.Now().UnixNano())
			tx, _ := entities.NewTransfer(sender.ID, receiver.ID, int64((i+1)*100), key, "list test")
			tx.Complete()
			require.NoError(t, ds.Insert(context.Background(), tx))
		}

		list, err := ds.SelectListByUserID(context.Background(), sender.ID, 0, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 5)
	})

	t.Run("ページネーション", func(t *testing.T) {
		list, err := ds.SelectListByUserID(context.Background(), sender.ID, 0, 3)
		require.NoError(t, err)
		assert.Len(t, list, 3)

		list2, err := ds.SelectListByUserID(context.Background(), sender.ID, 3, 3)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list2), 2)
	})
}

// ========================================
// TransactionDataSource Count Tests
// ========================================

func TestTransactionDataSource_CountByUserID(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewTransactionDataSource(db)
	sender := createTestUser(t, db, "count_sender")
	receiver := createTestUser(t, db, "count_receiver")

	t.Run("ユーザーのトランザクション数を取得", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			key := fmt.Sprintf("count-key-%d-%d", i, time.Now().UnixNano())
			tx, _ := entities.NewTransfer(sender.ID, receiver.ID, 100, key, "count test")
			tx.Complete()
			require.NoError(t, ds.Insert(context.Background(), tx))
		}

		count, err := ds.CountByUserID(context.Background(), sender.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))
	})
}

// ========================================
// TransactionDataSource Update Tests
// ========================================

func TestTransactionDataSource_Update(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewTransactionDataSource(db)
	sender := createTestUser(t, db, "update_sender")
	receiver := createTestUser(t, db, "update_receiver")

	t.Run("トランザクションステータスを更新", func(t *testing.T) {
		key := fmt.Sprintf("update-key-%d", time.Now().UnixNano())
		tx, _ := entities.NewTransfer(sender.ID, receiver.ID, 1000, key, "Update test")
		require.NoError(t, ds.Insert(context.Background(), tx))

		// ステータスをcompletedに変更
		tx.Complete()
		err := ds.Update(context.Background(), tx)
		require.NoError(t, err)

		retrieved, err := ds.Select(context.Background(), tx.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.TransactionStatusCompleted, retrieved.Status)
	})
}
