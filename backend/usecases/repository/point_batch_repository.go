package repository

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// PointBatchRepository はポイントバッチのリポジトリインターフェース
type PointBatchRepository interface {
	// Create は新しいポイントバッチを作成
	Create(ctx context.Context, batch *entities.PointBatch) error

	// ConsumePointsFIFO は古いバッチから順にポイントを消費（FIFO）
	ConsumePointsFIFO(ctx context.Context, userID uuid.UUID, amount int64) error

	// FindExpiredBatches は期限切れで残量があるバッチを検索
	FindExpiredBatches(ctx context.Context, before time.Time, limit int) ([]*entities.PointBatch, error)

	// MarkExpired はバッチを失効済みに更新（remaining_amount = 0）
	MarkExpired(ctx context.Context, batchID uuid.UUID) error

	// FindUpcomingExpirations はユーザーの有効なバッチを期限が近い順に取得
	FindUpcomingExpirations(ctx context.Context, userID uuid.UUID) ([]*entities.PointBatch, error)
}
