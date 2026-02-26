//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"

	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// AnalyticsDataSource Tests
// ========================================

func TestAnalyticsDataSource_GetUserBalanceSummary(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewAnalyticsDataSource(db)

	t.Run("アクティブユーザーの残高サマリーを取得", func(t *testing.T) {
		// テストユーザーを作成
		createTestUserWithBalanceDB(t, db, "analytics_user1", 5000)
		createTestUserWithBalanceDB(t, db, "analytics_user2", 3000)
		createTestUserWithBalanceDB(t, db, "analytics_user3", 2000)

		result, err := ds.GetUserBalanceSummary(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.GreaterOrEqual(t, result.TotalBalance, int64(10000))
		assert.GreaterOrEqual(t, result.ActiveUsers, int64(3))
		assert.Greater(t, result.AverageBalance, float64(0))
	})

	t.Run("ユーザーがいない場合もエラーにならない", func(t *testing.T) {
		// 空のDBで実行（前のテストのデータが残っている場合があるので、結果は0以上を確認）
		result, err := ds.GetUserBalanceSummary(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestAnalyticsDataSource_GetTopHolders(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewAnalyticsDataSource(db)

	t.Run("ポイント保有上位ユーザーを取得", func(t *testing.T) {
		// 異なる残高のユーザーを作成
		createTestUserWithBalanceDB(t, db, "top_user1", 10000)
		createTestUserWithBalanceDB(t, db, "top_user2", 5000)
		createTestUserWithBalanceDB(t, db, "top_user3", 8000)
		createTestUserWithBalanceDB(t, db, "top_user4", 3000)
		createTestUserWithBalanceDB(t, db, "top_user5", 15000)

		holders, err := ds.GetTopHolders(context.Background(), 3)
		require.NoError(t, err)
		assert.Len(t, holders, 3)

		// 降順であることを確認
		for i := 0; i < len(holders)-1; i++ {
			assert.GreaterOrEqual(t, holders[i].Balance, holders[i+1].Balance)
		}
	})
}

func TestAnalyticsDataSource_GetMonthlyTransactionCount(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewAnalyticsDataSource(db)

	t.Run("今月のトランザクション数を取得", func(t *testing.T) {
		count, err := ds.GetMonthlyTransactionCount(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

func TestAnalyticsDataSource_GetMonthlyIssuedPoints(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewAnalyticsDataSource(db)

	t.Run("今月の発行ポイント数を取得", func(t *testing.T) {
		total, err := ds.GetMonthlyIssuedPoints(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(0))
	})
}

func TestAnalyticsDataSource_GetTransactionTypeBreakdown(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewAnalyticsDataSource(db)

	t.Run("トランザクション種別構成を取得", func(t *testing.T) {
		breakdowns, err := ds.GetTransactionTypeBreakdown(context.Background())
		require.NoError(t, err)
		// データがなくてもエラーにならない
		assert.NotNil(t, breakdowns)
	})
}
