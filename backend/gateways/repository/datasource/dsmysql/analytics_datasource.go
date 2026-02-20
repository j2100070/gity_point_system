package dsmysql

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
)

// AnalyticsDataSource は分析用データソースインターフェース
type AnalyticsDataSource interface {
	// GetUserBalanceSummary はアクティブユーザーの残高サマリーを取得
	GetUserBalanceSummary(ctx context.Context) (*entities.AnalyticsSummaryResult, error)

	// GetTopHolders はポイント保有上位ユーザーを取得
	GetTopHolders(ctx context.Context, limit int) ([]*entities.TopHolderResult, error)

	// GetDailyStats は日別統計を取得
	GetDailyStats(ctx context.Context, since time.Time) ([]*entities.DailyStatResult, error)

	// GetTransactionTypeBreakdown はトランザクション種別構成を取得
	GetTransactionTypeBreakdown(ctx context.Context) ([]*entities.TypeBreakdownResult, error)

	// GetMonthlyIssuedPoints は今月の発行ポイント数を取得
	GetMonthlyIssuedPoints(ctx context.Context) (int64, error)

	// GetMonthlyTransactionCount は今月のトランザクション数を取得
	GetMonthlyTransactionCount(ctx context.Context) (int64, error)
}
