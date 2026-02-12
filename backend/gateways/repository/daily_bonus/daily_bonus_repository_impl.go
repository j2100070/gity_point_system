package daily_bonus

import (
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
func (r *DailyBonusRepositoryImpl) Create(bonus *entities.DailyBonus) error {
	return r.ds.Insert(bonus)
}

// Update はデイリーボーナスを更新
func (r *DailyBonusRepositoryImpl) Update(bonus *entities.DailyBonus) error {
	return r.ds.Update(bonus)
}

// Read はIDでデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) Read(id uuid.UUID) (*entities.DailyBonus, error) {
	return r.ds.Select(id)
}

// ReadByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) ReadByUserAndDate(userID uuid.UUID, date time.Time) (*entities.DailyBonus, error) {
	return r.ds.SelectByUserAndDate(userID, date)
}

// ReadRecentByUser はユーザーの最近のデイリーボーナスを取得
func (r *DailyBonusRepositoryImpl) ReadRecentByUser(userID uuid.UUID, limit int) ([]*entities.DailyBonus, error) {
	return r.ds.SelectRecentByUser(userID, limit)
}

// CountAllCompletedByUser はユーザーの全達成回数をカウント
func (r *DailyBonusRepositoryImpl) CountAllCompletedByUser(userID uuid.UUID) (int64, error) {
	return r.ds.CountAllCompletedByUser(userID)
}
