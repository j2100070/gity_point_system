package dspostgresimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/gity/point-system/entities"
	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PointBatchModel はポイントバッチのGORMモデル
type PointBatchModel struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null"`
	OriginalAmount      int64      `gorm:"not null"`
	RemainingAmount     int64      `gorm:"not null"`
	SourceType          string     `gorm:"type:varchar(50);not null"`
	SourceTransactionID *uuid.UUID `gorm:"type:uuid"`
	ExpiresAt           time.Time  `gorm:"type:timestamptz;not null"`
	CreatedAt           time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName はテーブル名を指定
func (PointBatchModel) TableName() string {
	return "point_batches"
}

// PointBatchDataSource はポイントバッチのデータソース
type PointBatchDataSource struct {
	db infrapostgres.DB
}

// NewPointBatchDataSource は新しいPointBatchDataSourceを作成
func NewPointBatchDataSource(db infrapostgres.DB) *PointBatchDataSource {
	return &PointBatchDataSource{db: db}
}

// toEntity はGORMモデルをエンティティに変換
func (ds *PointBatchDataSource) toEntity(model *PointBatchModel) *entities.PointBatch {
	return &entities.PointBatch{
		ID:                  model.ID,
		UserID:              model.UserID,
		OriginalAmount:      model.OriginalAmount,
		RemainingAmount:     model.RemainingAmount,
		SourceType:          entities.PointBatchSourceType(model.SourceType),
		SourceTransactionID: model.SourceTransactionID,
		ExpiresAt:           model.ExpiresAt,
		CreatedAt:           model.CreatedAt,
	}
}

// toModel はエンティティをGORMモデルに変換
func (ds *PointBatchDataSource) toModel(batch *entities.PointBatch) *PointBatchModel {
	return &PointBatchModel{
		ID:                  batch.ID,
		UserID:              batch.UserID,
		OriginalAmount:      batch.OriginalAmount,
		RemainingAmount:     batch.RemainingAmount,
		SourceType:          string(batch.SourceType),
		SourceTransactionID: batch.SourceTransactionID,
		ExpiresAt:           batch.ExpiresAt,
		CreatedAt:           batch.CreatedAt,
	}
}

// Insert はポイントバッチを挿入
func (ds *PointBatchDataSource) Insert(ctx context.Context, batch *entities.PointBatch) error {
	db := infrapostgres.GetDB(ctx, ds.db.GetDB())
	model := ds.toModel(batch)
	return db.Create(model).Error
}

// ConsumePointsFIFO は古いバッチから順にポイントを消費（FIFO）
// トランザクションコンテキスト内で呼ぶこと
func (ds *PointBatchDataSource) ConsumePointsFIFO(ctx context.Context, userID uuid.UUID, amount int64) error {
	db := infrapostgres.GetDB(ctx, ds.db.GetDB())

	// 有効なバッチを古い順に取得（期限内かつ残量あり）
	var batches []PointBatchModel
	err := db.Where("user_id = ? AND remaining_amount > 0 AND expires_at > NOW()", userID).
		Order("created_at ASC").
		Find(&batches).Error
	if err != nil {
		return fmt.Errorf("failed to find active batches: %w", err)
	}

	remaining := amount
	for _, batch := range batches {
		if remaining <= 0 {
			break
		}

		consume := batch.RemainingAmount
		if consume > remaining {
			consume = remaining
		}

		// remaining_amountを減算
		err := db.Model(&PointBatchModel{}).
			Where("id = ?", batch.ID).
			Update("remaining_amount", gorm.Expr("remaining_amount - ?", consume)).Error
		if err != nil {
			return fmt.Errorf("failed to consume batch %s: %w", batch.ID, err)
		}

		remaining -= consume
	}

	return nil
}

// SelectExpiredBatches は期限切れで残量があるバッチを検索
func (ds *PointBatchDataSource) SelectExpiredBatches(ctx context.Context, before time.Time, limit int) ([]*entities.PointBatch, error) {
	db := infrapostgres.GetDB(ctx, ds.db.GetDB())

	var models []PointBatchModel
	err := db.Where("expires_at < ? AND remaining_amount > 0", before).
		Order("expires_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	batches := make([]*entities.PointBatch, len(models))
	for i, model := range models {
		batches[i] = ds.toEntity(&model)
	}
	return batches, nil
}

// MarkExpired はバッチのremaining_amountを0に更新
func (ds *PointBatchDataSource) MarkExpired(ctx context.Context, batchID uuid.UUID) error {
	db := infrapostgres.GetDB(ctx, ds.db.GetDB())
	return db.Model(&PointBatchModel{}).
		Where("id = ?", batchID).
		Update("remaining_amount", 0).Error
}

// SelectUpcomingExpirations はユーザーの1ヶ月以内に失効するバッチを期限が近い順に取得
func (ds *PointBatchDataSource) SelectUpcomingExpirations(ctx context.Context, userID uuid.UUID) ([]*entities.PointBatch, error) {
	db := infrapostgres.GetDB(ctx, ds.db.GetDB())

	var models []PointBatchModel
	err := db.Where("user_id = ? AND remaining_amount > 0 AND expires_at > NOW() AND expires_at <= NOW() + INTERVAL '1 month'", userID).
		Order("expires_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	batches := make([]*entities.PointBatch, len(models))
	for i, model := range models {
		batches[i] = ds.toEntity(&model)
	}
	return batches, nil
}
