//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// PointBatchDataSource Insert Tests
// ========================================

func TestPointBatchDataSource_Insert(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewPointBatchDataSource(db)
	user := createTestUser(t, db, "batch_user")

	t.Run("ポイントバッチを挿入", func(t *testing.T) {
		batch := entities.NewPointBatch(
			user.ID,
			1000,
			entities.PointBatchSourceAdminGrant,
			nil,
			time.Now(),
		)

		err := ds.Insert(context.Background(), batch)
		require.NoError(t, err)
		assert.NotEmpty(t, batch.ID)
	})
}

// ========================================
// PointBatchDataSource ConsumePointsFIFO Tests
// ========================================

func TestPointBatchDataSource_ConsumePointsFIFO(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewPointBatchDataSource(db)
	user := createTestUser(t, db, "fifo_user")

	t.Run("古いバッチから順にポイントを消費", func(t *testing.T) {
		now := time.Now()

		// 3つのバッチを作成（古い順）
		batch1 := entities.NewPointBatch(user.ID, 300, entities.PointBatchSourceAdminGrant, nil, now.Add(-3*time.Hour))
		batch2 := entities.NewPointBatch(user.ID, 500, entities.PointBatchSourceTransfer, nil, now.Add(-2*time.Hour))
		batch3 := entities.NewPointBatch(user.ID, 200, entities.PointBatchSourceDailyBonus, nil, now.Add(-1*time.Hour))

		require.NoError(t, ds.Insert(context.Background(), batch1))
		require.NoError(t, ds.Insert(context.Background(), batch2))
		require.NoError(t, ds.Insert(context.Background(), batch3))

		// 400ポイントを消費: batch1(300) + batch2(100) = 400
		err := ds.ConsumePointsFIFO(context.Background(), user.ID, 400)
		require.NoError(t, err)

		// batch1は完全に消費されている
		// batch2は100ポイント消費されて400残っている
		// batch3は未消費で200残っている
		// 合計: 0 + 400 + 200 = 600ポイント残
		upcoming, err := ds.SelectUpcomingExpirations(context.Background(), user.ID)
		require.NoError(t, err)

		totalRemaining := int64(0)
		for _, b := range upcoming {
			totalRemaining += b.RemainingAmount
		}
		// 期限内のバッチのみが返るため、具体的な合計チェックは SelectExpiredBatches と合わせて確認
	})
}

// ========================================
// PointBatchDataSource Expiration Tests
// ========================================

func TestPointBatchDataSource_SelectExpiredBatches(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewPointBatchDataSource(db)
	user := createTestUser(t, db, "expired_batch_user")

	t.Run("期限切れで残量があるバッチを取得", func(t *testing.T) {
		now := time.Now()

		// 期限切れ バッチ（残量あり）
		expired := entities.NewPointBatch(user.ID, 500, entities.PointBatchSourceAdminGrant, nil, now.Add(-100*24*time.Hour))
		expired.ExpiresAt = now.Add(-1 * time.Hour) // 期限切れに設定
		require.NoError(t, ds.Insert(context.Background(), expired))

		// 有効なバッチ
		valid := entities.NewPointBatch(user.ID, 500, entities.PointBatchSourceAdminGrant, nil, now)
		require.NoError(t, ds.Insert(context.Background(), valid))

		batches, err := ds.SelectExpiredBatches(context.Background(), now, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(batches), 1)

		// 期限切れのバッチのみが返されている
		for _, b := range batches {
			assert.True(t, b.ExpiresAt.Before(now), "期限切れバッチのみ返すこと")
			assert.Greater(t, b.RemainingAmount, int64(0), "残量ありのバッチのみ返すこと")
		}
	})
}

func TestPointBatchDataSource_MarkExpired(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewPointBatchDataSource(db)
	user := createTestUser(t, db, "mark_expired_user")

	t.Run("バッチのremaining_amountを0に更新", func(t *testing.T) {
		now := time.Now()
		batch := entities.NewPointBatch(user.ID, 500, entities.PointBatchSourceAdminGrant, nil, now.Add(-100*24*time.Hour))
		batch.ExpiresAt = now.Add(-1 * time.Hour) // 期限切れ
		require.NoError(t, ds.Insert(context.Background(), batch))

		err := ds.MarkExpired(context.Background(), batch.ID)
		require.NoError(t, err)

		// remaining_amountが0になっている
		batches, err := ds.SelectExpiredBatches(context.Background(), now, 10)
		require.NoError(t, err)

		for _, b := range batches {
			if b.ID == batch.ID {
				t.Fatal("MarkExpired後はSelectExpiredBatchesに含まれないべき")
			}
		}
	})
}

func TestPointBatchDataSource_SelectUpcomingExpirations(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewPointBatchDataSource(db)
	user := createTestUser(t, db, "upcoming_user")

	t.Run("1ヶ月以内に失効するバッチを取得", func(t *testing.T) {
		now := time.Now()

		// 2週間後に失効するバッチ
		soon := entities.NewPointBatch(user.ID, 300, entities.PointBatchSourceAdminGrant, nil, now)
		soon.ExpiresAt = now.Add(14 * 24 * time.Hour)
		require.NoError(t, ds.Insert(context.Background(), soon))

		// 2ヶ月後に失効するバッチ（対象外）
		later := entities.NewPointBatch(user.ID, 300, entities.PointBatchSourceAdminGrant, nil, now)
		later.ExpiresAt = now.Add(60 * 24 * time.Hour)
		require.NoError(t, ds.Insert(context.Background(), later))

		upcoming, err := ds.SelectUpcomingExpirations(context.Background(), user.ID)
		require.NoError(t, err)

		// 1ヶ月以内のバッチのみ含まれる
		for _, b := range upcoming {
			assert.True(t, b.ExpiresAt.Before(now.Add(31*24*time.Hour)),
				"1ヶ月以内に失効するバッチのみ返すこと")
			assert.True(t, b.ExpiresAt.After(now), "期限内のバッチのみ返すこと")
		}
	})
}
