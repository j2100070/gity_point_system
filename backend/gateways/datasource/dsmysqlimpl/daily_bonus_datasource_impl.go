package dsmysqlimpl

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DailyBonusModel はAkerun入退室ベースのデイリーボーナスGORMモデル
type DailyBonusModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	BonusDate      time.Time  `gorm:"type:date;not null"`
	BonusPoints    int64      `gorm:"default:5;not null"`
	AkerunAccessID *string    `gorm:"type:text"`
	AkerunUserName *string    `gorm:"type:text"`
	AccessedAt     *time.Time `gorm:"type:timestamptz"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName はテーブル名を指定
func (DailyBonusModel) TableName() string {
	return "daily_bonuses"
}

// AkerunPollStateModel はポーリング状態のGORMモデル
type AkerunPollStateModel struct {
	ID           int       `gorm:"primary_key;default:1"`
	LastPolledAt time.Time `gorm:"type:timestamptz;not null"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName はテーブル名を指定
func (AkerunPollStateModel) TableName() string {
	return "akerun_poll_state"
}

// DailyBonusDataSource はデイリーボーナスのデータソース
type DailyBonusDataSource struct {
	db inframysql.DB
}

// NewDailyBonusDataSource は新しいDailyBonusDataSourceを作成
func NewDailyBonusDataSource(db inframysql.DB) *DailyBonusDataSource {
	return &DailyBonusDataSource{db: db}
}

// toEntity はGORMモデルをエンティティに変換
func (ds *DailyBonusDataSource) toEntity(model *DailyBonusModel) *entities.DailyBonus {
	bonus := &entities.DailyBonus{
		ID:          model.ID,
		UserID:      model.UserID,
		BonusDate:   model.BonusDate,
		BonusPoints: model.BonusPoints,
		AccessedAt:  model.AccessedAt,
		CreatedAt:   model.CreatedAt,
	}
	if model.AkerunAccessID != nil {
		bonus.AkerunAccessID = *model.AkerunAccessID
	}
	if model.AkerunUserName != nil {
		bonus.AkerunUserName = *model.AkerunUserName
	}
	return bonus
}

// toModel はエンティティをGORMモデルに変換
func (ds *DailyBonusDataSource) toModel(bonus *entities.DailyBonus) *DailyBonusModel {
	model := &DailyBonusModel{
		ID:          bonus.ID,
		UserID:      bonus.UserID,
		BonusDate:   bonus.BonusDate,
		BonusPoints: bonus.BonusPoints,
		AccessedAt:  bonus.AccessedAt,
		CreatedAt:   bonus.CreatedAt,
	}
	if bonus.AkerunAccessID != "" {
		model.AkerunAccessID = &bonus.AkerunAccessID
	}
	if bonus.AkerunUserName != "" {
		model.AkerunUserName = &bonus.AkerunUserName
	}
	return model
}

// Insert はデイリーボーナスを挿入
func (ds *DailyBonusDataSource) Insert(ctx context.Context, bonus *entities.DailyBonus) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := ds.toModel(bonus)
	return db.Create(model).Error
}

// SelectByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
func (ds *DailyBonusDataSource) SelectByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailyBonus, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model DailyBonusModel
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	err := db.Where("user_id = ? AND bonus_date = ?", userID, dateOnly).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return ds.toEntity(&model), nil
}

// SelectRecentByUser はユーザーの最近のデイリーボーナスを取得
func (ds *DailyBonusDataSource) SelectRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var models []DailyBonusModel
	err := db.
		Where("user_id = ?", userID).
		Order("bonus_date DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	bonuses := make([]*entities.DailyBonus, len(models))
	for i, model := range models {
		bonuses[i] = ds.toEntity(&model)
	}
	return bonuses, nil
}

// CountByUser はユーザーのボーナス獲得日数をカウント
func (ds *DailyBonusDataSource) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var count int64
	err := db.Model(&DailyBonusModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetLastPolledAt は前回ポーリング時刻を取得
func (ds *DailyBonusDataSource) GetLastPolledAt(ctx context.Context) (time.Time, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model AkerunPollStateModel
	err := db.First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Now().Add(-24 * time.Hour), nil // デフォルト: 24時間前
		}
		return time.Time{}, err
	}
	return model.LastPolledAt, nil
}

// UpdateLastPolledAt はポーリング時刻を更新
func (ds *DailyBonusDataSource) UpdateLastPolledAt(ctx context.Context, t time.Time) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	return db.Model(&AkerunPollStateModel{}).
		Where("id = 1").
		Updates(map[string]interface{}{
			"last_polled_at": t,
			"updated_at":     time.Now(),
		}).Error
}
