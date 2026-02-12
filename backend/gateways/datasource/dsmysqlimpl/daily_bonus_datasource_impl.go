package dsmysqlimpl

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DailyBonusModel はデイリーボーナスのGORMモデル
type DailyBonusModel struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null;index:idx_daily_bonuses_user_date"`
	BonusDate           time.Time  `gorm:"type:date;not null;index:idx_daily_bonuses_user_date"`
	LoginCompleted      bool       `gorm:"default:false;not null"`
	LoginCompletedAt    *time.Time `gorm:"type:timestamptz"`
	TransferCompleted   bool       `gorm:"default:false;not null"`
	TransferCompletedAt *time.Time `gorm:"type:timestamptz"`
	TransferTxID        *uuid.UUID `gorm:"type:uuid;column:transfer_transaction_id"`
	ExchangeCompleted   bool       `gorm:"default:false;not null"`
	ExchangeCompletedAt *time.Time `gorm:"type:timestamptz"`
	ExchangeID          *uuid.UUID `gorm:"type:uuid"`
	AllCompleted        bool       `gorm:"default:false;not null"`
	AllCompletedAt      *time.Time `gorm:"type:timestamptz"`
	TotalBonusPoints    int64      `gorm:"default:0;not null"`
	CreatedAt           time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName はテーブル名を指定
func (DailyBonusModel) TableName() string {
	return "daily_bonuses"
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
	return &entities.DailyBonus{
		ID:                  model.ID,
		UserID:              model.UserID,
		BonusDate:           model.BonusDate,
		LoginCompleted:      model.LoginCompleted,
		LoginCompletedAt:    model.LoginCompletedAt,
		TransferCompleted:   model.TransferCompleted,
		TransferCompletedAt: model.TransferCompletedAt,
		TransferTxID:        model.TransferTxID,
		ExchangeCompleted:   model.ExchangeCompleted,
		ExchangeCompletedAt: model.ExchangeCompletedAt,
		ExchangeID:          model.ExchangeID,
		AllCompleted:        model.AllCompleted,
		AllCompletedAt:      model.AllCompletedAt,
		TotalBonusPoints:    model.TotalBonusPoints,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

// toModel はエンティティをGORMモデルに変換
func (ds *DailyBonusDataSource) toModel(bonus *entities.DailyBonus) *DailyBonusModel {
	return &DailyBonusModel{
		ID:                  bonus.ID,
		UserID:              bonus.UserID,
		BonusDate:           bonus.BonusDate,
		LoginCompleted:      bonus.LoginCompleted,
		LoginCompletedAt:    bonus.LoginCompletedAt,
		TransferCompleted:   bonus.TransferCompleted,
		TransferCompletedAt: bonus.TransferCompletedAt,
		TransferTxID:        bonus.TransferTxID,
		ExchangeCompleted:   bonus.ExchangeCompleted,
		ExchangeCompletedAt: bonus.ExchangeCompletedAt,
		ExchangeID:          bonus.ExchangeID,
		AllCompleted:        bonus.AllCompleted,
		AllCompletedAt:      bonus.AllCompletedAt,
		TotalBonusPoints:    bonus.TotalBonusPoints,
		CreatedAt:           bonus.CreatedAt,
		UpdatedAt:           bonus.UpdatedAt,
	}
}

// Insert はデイリーボーナスを挿入
func (ds *DailyBonusDataSource) Insert(ctx context.Context, bonus *entities.DailyBonus) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := ds.toModel(bonus)
	return db.Create(model).Error
}

// Update はデイリーボーナスを更新
func (ds *DailyBonusDataSource) Update(ctx context.Context, bonus *entities.DailyBonus) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := ds.toModel(bonus)
	return db.Save(model).Error
}

// Select はIDでデイリーボーナスを取得
func (ds *DailyBonusDataSource) Select(ctx context.Context, id uuid.UUID) (*entities.DailyBonus, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model DailyBonusModel
	err := db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return ds.toEntity(&model), nil
}

// SelectByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
func (ds *DailyBonusDataSource) SelectByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailyBonus, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model DailyBonusModel
	// 日付のみで比較（時刻は無視）
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

// CountAllCompletedByUser はユーザーの全達成回数をカウント
func (ds *DailyBonusDataSource) CountAllCompletedByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var count int64
	err := db.Model(&DailyBonusModel{}).
		Where("user_id = ? AND all_completed = ?", userID, true).
		Count(&count).Error
	return count, err
}
