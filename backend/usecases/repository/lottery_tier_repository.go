package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// LotteryTierRepository はボーナス抽選ティアのリポジトリインターフェース
type LotteryTierRepository interface {
	// ReadAll は全ティアを取得（display_order順）
	ReadAll(ctx context.Context) ([]*entities.LotteryTier, error)

	// ReadActive はアクティブなティアのみ取得（display_order順）
	ReadActive(ctx context.Context) ([]*entities.LotteryTier, error)

	// Create はティアを作成
	Create(ctx context.Context, tier *entities.LotteryTier) error

	// Update はティアを更新
	Update(ctx context.Context, tier *entities.LotteryTier) error

	// Delete はティアを削除
	Delete(ctx context.Context, id uuid.UUID) error

	// ReplaceAll は全ティアを一括置換（トランザクション内で使用）
	ReplaceAll(ctx context.Context, tiers []*entities.LotteryTier) error
}
