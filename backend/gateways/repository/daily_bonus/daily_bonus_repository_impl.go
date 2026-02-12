package daily_bonus

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/google/uuid"
)

// DailyBonusRepositoryImpl はデイリーボーナスリポジトリの実装
type DailyBonusRepositoryImpl struct {
	ds *dsmysqlimpl.DailyBonusDataSource
}

// NewDailyBonusRepository は新しいDailyBonusRepositoryを作成
func NewDailyBonusRepository(ds *dsmysqlimpl.DailyBonusDataSource) *DailyBonusRepositoryImpl {
	return &DailyBonusRepositoryImpl{ds: ds}
}

// Create はデイリーボーナスを作成
func (r *DailyBonusRepositoryImpl) Create(ctx context.Context, bonus *entities.DailyBonus) error {
	return r.ds.Insert(ctx, bonus)
}

// Update はデイリーボーナスを更新
func (r *DailyBonusRepositoryImpl) Update(ctx context.Context, bonus *entities.DailyBonus) error {
	return r.ds.Update(ctx, bonus)
}

// Read はIDでデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.DailyBonus, error) {
	return r.ds.Select(ctx, id)
}

// ReadByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) ReadByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailyBonus, error) {
	return r.ds.SelectByUserAndDate(ctx, userID, date)
}

// ReadRecentByUser はユーザーの最近のデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) ReadRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error) {
	return r.ds.SelectRecentByUser(ctx, userID, limit)
}

// CountAllCompletedByUser はユーザーの全達成回数をカウント
func (r *DailyBonusRepositoryImpl) CountAllCompletedByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.ds.CountAllCompletedByUser(ctx, userID)
}
