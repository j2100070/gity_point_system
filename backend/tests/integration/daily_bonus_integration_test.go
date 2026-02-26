//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDailyBonus(t *testing.T) (inputport.DailyBonusInputPort, infrapostgres.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := infrapostgres.NewGormTransactionManager(db.GetDB())

	dailyBonus := interactor.NewDailyBonusInteractor(
		repos.DailyBonus, repos.User, repos.Transaction, txManager, repos.SystemSettings, repos.PointBatch, repos.LotteryTier, lg,
	)
	return dailyBonus, db
}

// seedLotteryTier はテスト用の抽選ティアをDBに挿入する
func seedLotteryTier(t *testing.T, db infrapostgres.DB) {
	t.Helper()
	err := db.GetDB().Exec(`
		INSERT INTO bonus_lottery_tiers (id, name, points, probability, display_order, is_active)
		VALUES (?, '通常', 5, 100.00, 1, true)
		ON CONFLICT DO NOTHING
	`, uuid.New()).Error
	require.NoError(t, err)
}

// seedPendingBonus は対象ユーザーの今日の未抽選ボーナスレコードを挿入する
func seedPendingBonus(t *testing.T, db infrapostgres.DB, userID uuid.UUID) {
	t.Helper()
	bonusDate := entities.GetBonusDateJST(time.Now())
	now := time.Now()
	err := db.GetDB().Exec(`
		INSERT INTO daily_bonuses (id, user_id, bonus_date, bonus_points, akerun_user_name, accessed_at, is_drawn, is_viewed, created_at)
		VALUES (?, ?, ?, 5, 'test_user', ?, false, false, ?)
	`, uuid.New(), userID, bonusDate, now, now).Error
	require.NoError(t, err)
}

// TestDailyBonus_DrawLottery はルーレット実行→ポイント付与を検証
func TestDailyBonus_DrawLottery(t *testing.T) {
	dailyBonus, db := setupDailyBonus(t)
	ctx := context.Background()

	user := createTestUser(t, db, "daily_bonus_user")
	seedLotteryTier(t, db)
	seedPendingBonus(t, db, user.ID)

	resp, err := dailyBonus.DrawLotteryAndGrant(ctx, &inputport.DrawLotteryRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.GreaterOrEqual(t, resp.BonusPoints, int64(0))
}

// TestDailyBonus_DuplicateDraw は同日の重複ルーレットを検証（2回目は既存結果を返す）
func TestDailyBonus_DuplicateDraw(t *testing.T) {
	dailyBonus, db := setupDailyBonus(t)
	ctx := context.Background()

	user := createTestUser(t, db, "daily_dup_user")
	seedLotteryTier(t, db)
	seedPendingBonus(t, db, user.ID)

	// 1回目
	resp1, err := dailyBonus.DrawLotteryAndGrant(ctx, &inputport.DrawLotteryRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp1)

	// 2回目（既に抽選済み → 同じ結果を返すはず）
	resp2, err := dailyBonus.DrawLotteryAndGrant(ctx, &inputport.DrawLotteryRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, resp1.BonusPoints, resp2.BonusPoints)
}

// TestDailyBonus_GetTodayBonus はボーナス取得状態を検証
func TestDailyBonus_GetTodayBonus(t *testing.T) {
	dailyBonus, db := setupDailyBonus(t)
	ctx := context.Background()

	user := createTestUser(t, db, "daily_status_user")
	seedLotteryTier(t, db)

	// 未獲得の状態
	status, err := dailyBonus.GetTodayBonus(ctx, &inputport.GetTodayBonusRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)
	assert.Nil(t, status.DailyBonus, "未獲得の場合はnilであること")

	// ボーナスレコードを挿入
	seedPendingBonus(t, db, user.ID)

	// 未抽選状態
	status, err = dailyBonus.GetTodayBonus(ctx, &inputport.GetTodayBonusRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, status.DailyBonus, "未抽選ボーナスが返ること")
	assert.False(t, status.DailyBonus.IsDrawn, "is_drawn=false であること")

	// ルーレット実行
	_, err = dailyBonus.DrawLotteryAndGrant(ctx, &inputport.DrawLotteryRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)

	// 獲得後
	status, err = dailyBonus.GetTodayBonus(ctx, &inputport.GetTodayBonusRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, status.DailyBonus, "獲得後もDailyBonusが返ること")
	assert.True(t, status.DailyBonus.IsDrawn, "is_drawn=true であること")
}
