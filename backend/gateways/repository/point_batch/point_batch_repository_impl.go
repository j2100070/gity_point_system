package point_batch

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/google/uuid"
)

// PointBatchRepositoryImpl はポイントバッチリポジトリの実装
type PointBatchRepositoryImpl struct {
	ds *dsmysqlimpl.PointBatchDataSource
}

// NewPointBatchRepository は新しいPointBatchRepositoryを作成
func NewPointBatchRepository(ds *dsmysqlimpl.PointBatchDataSource) *PointBatchRepositoryImpl {
	return &PointBatchRepositoryImpl{ds: ds}
}

// Create は新しいポイントバッチを作成
func (r *PointBatchRepositoryImpl) Create(ctx context.Context, batch *entities.PointBatch) error {
	return r.ds.Insert(ctx, batch)
}

// ConsumePointsFIFO は古いバッチから順にポイントを消費（FIFO）
func (r *PointBatchRepositoryImpl) ConsumePointsFIFO(ctx context.Context, userID uuid.UUID, amount int64) error {
	return r.ds.ConsumePointsFIFO(ctx, userID, amount)
}

// FindExpiredBatches は期限切れで残量があるバッチを検索
func (r *PointBatchRepositoryImpl) FindExpiredBatches(ctx context.Context, before time.Time, limit int) ([]*entities.PointBatch, error) {
	return r.ds.SelectExpiredBatches(ctx, before, limit)
}

// MarkExpired はバッチを失効済みに更新
func (r *PointBatchRepositoryImpl) MarkExpired(ctx context.Context, batchID uuid.UUID) error {
	return r.ds.MarkExpired(ctx, batchID)
}

// FindUpcomingExpirations はユーザーの有効なバッチを期限が近い順に取得
func (r *PointBatchRepositoryImpl) FindUpcomingExpirations(ctx context.Context, userID uuid.UUID) ([]*entities.PointBatch, error) {
	return r.ds.SelectUpcomingExpirations(ctx, userID)
}
