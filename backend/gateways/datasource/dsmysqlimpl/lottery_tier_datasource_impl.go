package dsmysqlimpl

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
)

// LotteryTierModel はボーナス抽選ティアのGORMモデル
type LotteryTierModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string    `gorm:"type:varchar(50);not null"`
	Points       int64     `gorm:"not null;default:0"`
	Probability  float64   `gorm:"type:decimal(5,2);not null;default:0"`
	DisplayOrder int       `gorm:"not null;default:0"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName はテーブル名を指定
func (LotteryTierModel) TableName() string {
	return "bonus_lottery_tiers"
}

// LotteryTierDataSource はボーナス抽選ティアのデータソース
type LotteryTierDataSource struct {
	db inframysql.DB
}

// NewLotteryTierDataSource は新しいLotteryTierDataSourceを作成
func NewLotteryTierDataSource(db inframysql.DB) *LotteryTierDataSource {
	return &LotteryTierDataSource{db: db}
}

func (ds *LotteryTierDataSource) toEntity(model *LotteryTierModel) *entities.LotteryTier {
	return &entities.LotteryTier{
		ID:           model.ID,
		Name:         model.Name,
		Points:       model.Points,
		Probability:  model.Probability,
		DisplayOrder: model.DisplayOrder,
		IsActive:     model.IsActive,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func (ds *LotteryTierDataSource) toModel(tier *entities.LotteryTier) *LotteryTierModel {
	return &LotteryTierModel{
		ID:           tier.ID,
		Name:         tier.Name,
		Points:       tier.Points,
		Probability:  tier.Probability,
		DisplayOrder: tier.DisplayOrder,
		IsActive:     tier.IsActive,
		CreatedAt:    tier.CreatedAt,
		UpdatedAt:    tier.UpdatedAt,
	}
}

// SelectAll は全ティアを取得（display_order順）
func (ds *LotteryTierDataSource) SelectAll(ctx context.Context) ([]*entities.LotteryTier, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var models []LotteryTierModel
	err := db.Order("display_order ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	tiers := make([]*entities.LotteryTier, len(models))
	for i, model := range models {
		tiers[i] = ds.toEntity(&model)
	}
	return tiers, nil
}

// SelectActive はアクティブなティアのみ取得
func (ds *LotteryTierDataSource) SelectActive(ctx context.Context) ([]*entities.LotteryTier, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var models []LotteryTierModel
	err := db.Where("is_active = ?", true).Order("display_order ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	tiers := make([]*entities.LotteryTier, len(models))
	for i, model := range models {
		tiers[i] = ds.toEntity(&model)
	}
	return tiers, nil
}

// Insert はティアを作成
func (ds *LotteryTierDataSource) Insert(ctx context.Context, tier *entities.LotteryTier) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := ds.toModel(tier)
	return db.Create(model).Error
}

// Update はティアを更新
func (ds *LotteryTierDataSource) Update(ctx context.Context, tier *entities.LotteryTier) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := ds.toModel(tier)
	return db.Save(model).Error
}

// Delete はティアを削除
func (ds *LotteryTierDataSource) Delete(ctx context.Context, id uuid.UUID) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	return db.Delete(&LotteryTierModel{}, "id = ?", id).Error
}

// ReplaceAll は全ティアを一括置換
func (ds *LotteryTierDataSource) ReplaceAll(ctx context.Context, tiers []*entities.LotteryTier) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	// 既存のティアを全削除
	if err := db.Where("1 = 1").Delete(&LotteryTierModel{}).Error; err != nil {
		return err
	}

	// 新しいティアを挿入
	for _, tier := range tiers {
		model := ds.toModel(tier)
		if err := db.Create(model).Error; err != nil {
			return err
		}
	}

	return nil
}
