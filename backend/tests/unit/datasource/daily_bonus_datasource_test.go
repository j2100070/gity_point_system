//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// DailyBonusDataSource Insert / Select Tests
// ========================================

func TestDailyBonusDataSource_InsertAndSelectByUserAndDate(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewDailyBonusDataSource(db)
	user := createTestUser(t, db, "bonus_user")

	t.Run("デイリーボーナスを作成してユーザー+日付で取得", func(t *testing.T) {
		today := time.Now()
		bonus := &entities.DailyBonus{
			ID:             uuid.New(),
			UserID:         user.ID,
			BonusDate:      time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()),
			BonusPoints:    5,
			AkerunAccessID: "access-123",
			AkerunUserName: "Test User",
			IsViewed:       false,
			IsDrawn:        false,
			CreatedAt:      time.Now(),
		}

		err := ds.Insert(context.Background(), bonus)
		require.NoError(t, err)

		retrieved, err := ds.SelectByUserAndDate(context.Background(), user.ID, today)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, bonus.ID, retrieved.ID)
		assert.Equal(t, int64(5), retrieved.BonusPoints)
		assert.False(t, retrieved.IsViewed)
		assert.False(t, retrieved.IsDrawn)
	})

	t.Run("ボーナスがない日はnilを返す", func(t *testing.T) {
		yesterday := time.Now().Add(-48 * time.Hour)
		retrieved, err := ds.SelectByUserAndDate(context.Background(), user.ID, yesterday)
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

// ========================================
// DailyBonusDataSource List Tests
// ========================================

func TestDailyBonusDataSource_SelectRecentByUser(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewDailyBonusDataSource(db)
	user := createTestUser(t, db, "recent_bonus_user")

	t.Run("最近のデイリーボーナスを取得", func(t *testing.T) {
		now := time.Now()
		for i := 0; i < 5; i++ {
			date := now.AddDate(0, 0, -i)
			bonus := &entities.DailyBonus{
				ID:          uuid.New(),
				UserID:      user.ID,
				BonusDate:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()),
				BonusPoints: int64(5 + i),
				IsViewed:    false,
				IsDrawn:     false,
				CreatedAt:   date,
			}
			require.NoError(t, ds.Insert(context.Background(), bonus))
		}

		recent, err := ds.SelectRecentByUser(context.Background(), user.ID, 3)
		require.NoError(t, err)
		assert.Len(t, recent, 3)

		// bonus_date降順であることを確認
		for i := 0; i < len(recent)-1; i++ {
			assert.True(t, recent[i].BonusDate.After(recent[i+1].BonusDate) ||
				recent[i].BonusDate.Equal(recent[i+1].BonusDate))
		}
	})
}

// ========================================
// DailyBonusDataSource Count Tests
// ========================================

func TestDailyBonusDataSource_CountByUser(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewDailyBonusDataSource(db)
	user := createTestUser(t, db, "count_bonus_user")

	t.Run("ユーザーのボーナス獲得日数を取得", func(t *testing.T) {
		now := time.Now()
		for i := 0; i < 3; i++ {
			date := now.AddDate(0, 0, -i)
			bonus := &entities.DailyBonus{
				ID:          uuid.New(),
				UserID:      user.ID,
				BonusDate:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()),
				BonusPoints: 5,
				IsViewed:    false,
				IsDrawn:     false,
				CreatedAt:   date,
			}
			require.NoError(t, ds.Insert(context.Background(), bonus))
		}

		count, err := ds.CountByUser(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

// ========================================
// DailyBonusDataSource Update Tests
// ========================================

func TestDailyBonusDataSource_UpdateIsViewed(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewDailyBonusDataSource(db)
	user := createTestUser(t, db, "viewed_bonus_user")

	t.Run("閲覧済みフラグを更新", func(t *testing.T) {
		today := time.Now()
		bonus := &entities.DailyBonus{
			ID:          uuid.New(),
			UserID:      user.ID,
			BonusDate:   time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()),
			BonusPoints: 5,
			IsViewed:    false,
			IsDrawn:     false,
			CreatedAt:   time.Now(),
		}
		require.NoError(t, ds.Insert(context.Background(), bonus))

		err := ds.UpdateIsViewed(context.Background(), bonus.ID)
		require.NoError(t, err)

		retrieved, err := ds.SelectByUserAndDate(context.Background(), user.ID, today)
		require.NoError(t, err)
		assert.True(t, retrieved.IsViewed)
	})
}

func TestDailyBonusDataSource_UpdateDrawnResult(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewDailyBonusDataSource(db)
	user := createTestUser(t, db, "drawn_bonus_user")

	t.Run("抽選結果を更新", func(t *testing.T) {
		today := time.Now()
		bonus := &entities.DailyBonus{
			ID:          uuid.New(),
			UserID:      user.ID,
			BonusDate:   time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()),
			BonusPoints: 5,
			IsViewed:    false,
			IsDrawn:     false,
			CreatedAt:   time.Now(),
		}
		require.NoError(t, ds.Insert(context.Background(), bonus))

		tierID := uuid.New()
		err := ds.UpdateDrawnResult(context.Background(), bonus.ID, 50, &tierID, "ゴールド")
		require.NoError(t, err)

		retrieved, err := ds.SelectByUserAndDate(context.Background(), user.ID, today)
		require.NoError(t, err)
		assert.True(t, retrieved.IsDrawn)
		assert.Equal(t, int64(50), retrieved.BonusPoints)
		assert.NotNil(t, retrieved.LotteryTierID)
		assert.Equal(t, tierID, *retrieved.LotteryTierID)
		assert.Equal(t, "ゴールド", retrieved.LotteryTierName)
	})

	t.Run("既に抽選済みの場合は更新されない", func(t *testing.T) {
		today := time.Now()
		date := today.AddDate(0, 0, -1) // 昨日
		bonus := &entities.DailyBonus{
			ID:          uuid.New(),
			UserID:      user.ID,
			BonusDate:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()),
			BonusPoints: 5,
			IsViewed:    false,
			IsDrawn:     true, // 既に抽選済み
			CreatedAt:   date,
		}
		require.NoError(t, ds.Insert(context.Background(), bonus))

		// is_drawn = false の条件で更新されないはず
		err := ds.UpdateDrawnResult(context.Background(), bonus.ID, 100, nil, "プラチナ")
		require.NoError(t, err)

		retrieved, err := ds.SelectByUserAndDate(context.Background(), user.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(5), retrieved.BonusPoints, "既に抽選済みなので更新されない")
	})
}

// ========================================
// AkerunPollState Tests
// ========================================

func TestDailyBonusDataSource_PollState(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewDailyBonusDataSource(db)

	t.Run("ポーリング時刻の取得と更新", func(t *testing.T) {
		// マイグレーションでシードデータが入っているので、まず削除して「レコードなし」状態をテスト
		db.GetDB().Exec("DELETE FROM akerun_poll_state")

		// レコードなしの場合は24時間前を返す
		lastPolled, err := ds.GetLastPolledAt(context.Background())
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now().Add(-24*time.Hour), lastPolled, 2*time.Second)

		// レコードを再挿入してUpdateをテスト
		oldTime := time.Now().Add(-1 * time.Hour)
		db.GetDB().Exec("INSERT INTO akerun_poll_state (id, last_polled_at, updated_at) VALUES (1, ?, ?)", oldTime, time.Now())

		// ポーリング時刻を更新
		newPolledAt := time.Now()
		err = ds.UpdateLastPolledAt(context.Background(), newPolledAt)
		require.NoError(t, err)

		// 更新後の取得
		lastPolled2, err := ds.GetLastPolledAt(context.Background())
		require.NoError(t, err)
		assert.WithinDuration(t, newPolledAt, lastPolled2, 2*time.Second)
	})
}
