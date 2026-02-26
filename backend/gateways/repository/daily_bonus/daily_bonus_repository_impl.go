package daily_bonus

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/google/uuid"
)

// DailyBonusRepositoryImpl はデイリーボーナスリポジトリの実装
type DailyBonusRepositoryImpl struct {
	ds *dspostgresimpl.DailyBonusDataSource
}

// NewDailyBonusRepository は新しいDailyBonusRepositoryを作成
func NewDailyBonusRepository(ds *dspostgresimpl.DailyBonusDataSource) *DailyBonusRepositoryImpl {
	return &DailyBonusRepositoryImpl{ds: ds}
}

// Create はデイリーボーナスを作成
func (r *DailyBonusRepositoryImpl) Create(ctx context.Context, bonus *entities.DailyBonus) error {
	return r.ds.Insert(ctx, bonus)
}

// ReadByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) ReadByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailyBonus, error) {
	return r.ds.SelectByUserAndDate(ctx, userID, date)
}

// ReadRecentByUser はユーザーの最近のデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) ReadRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error) {
	return r.ds.SelectRecentByUser(ctx, userID, limit)
}

// CountByUser はユーザーのボーナス獲得日数をカウント
func (r *DailyBonusRepositoryImpl) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.ds.CountByUser(ctx, userID)
}

// GetLastPolledAt は前回ポーリング時刻を取得
func (r *DailyBonusRepositoryImpl) GetLastPolledAt(ctx context.Context) (time.Time, error) {
	return r.ds.GetLastPolledAt(ctx)
}

// UpdateLastPolledAt はポーリング時刻を更新
func (r *DailyBonusRepositoryImpl) UpdateLastPolledAt(ctx context.Context, t time.Time) error {
	return r.ds.UpdateLastPolledAt(ctx, t)
}

// MarkAsViewed はデイリーボーナスを閲覧済みにする
func (r *DailyBonusRepositoryImpl) MarkAsViewed(ctx context.Context, id uuid.UUID) error {
	return r.ds.UpdateIsViewed(ctx, id)
}

// UpdateDrawnResult は抽選結果を更新する（ルーレット実行時）
func (r *DailyBonusRepositoryImpl) UpdateDrawnResult(ctx context.Context, id uuid.UUID, bonusPoints int64, lotteryTierID *uuid.UUID, lotteryTierName string) error {
	return r.ds.UpdateDrawnResult(ctx, id, bonusPoints, lotteryTierID, lotteryTierName)
}
