package lottery_tier

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/google/uuid"
)

// LotteryTierRepositoryImpl はボーナス抽選ティアリポジトリの実装
type LotteryTierRepositoryImpl struct {
	ds *dspostgresimpl.LotteryTierDataSource
}

// NewLotteryTierRepository は新しいLotteryTierRepositoryを作成
func NewLotteryTierRepository(ds *dspostgresimpl.LotteryTierDataSource) *LotteryTierRepositoryImpl {
	return &LotteryTierRepositoryImpl{ds: ds}
}

// ReadAll は全ティアを取得
func (r *LotteryTierRepositoryImpl) ReadAll(ctx context.Context) ([]*entities.LotteryTier, error) {
	return r.ds.SelectAll(ctx)
}

// ReadActive はアクティブなティアのみ取得
func (r *LotteryTierRepositoryImpl) ReadActive(ctx context.Context) ([]*entities.LotteryTier, error) {
	return r.ds.SelectActive(ctx)
}

// Create はティアを作成
func (r *LotteryTierRepositoryImpl) Create(ctx context.Context, tier *entities.LotteryTier) error {
	return r.ds.Insert(ctx, tier)
}

// Update はティアを更新
func (r *LotteryTierRepositoryImpl) Update(ctx context.Context, tier *entities.LotteryTier) error {
	return r.ds.Update(ctx, tier)
}

// Delete はティアを削除
func (r *LotteryTierRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.ds.Delete(ctx, id)
}

// ReplaceAll は全ティアを一括置換
func (r *LotteryTierRepositoryImpl) ReplaceAll(ctx context.Context, tiers []*entities.LotteryTier) error {
	return r.ds.ReplaceAll(ctx, tiers)
}
