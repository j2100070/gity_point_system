package dsmysql

import (
	"context"
	"time"
)

// AnalyticsSummaryResult は集約サマリーの結果
type AnalyticsSummaryResult struct {
	TotalBalance   int64
	AverageBalance float64
	ActiveUsers    int64
}

// DailyStatResult は日別統計の結果
type DailyStatResult struct {
	Date        time.Time
	Issued      int64
	Consumed    int64
	Transferred int64
}

// TypeBreakdownResult はトランザクション種別構成の結果
type TypeBreakdownResult struct {
	Type        string
	Count       int64
	TotalAmount int64
}

// TopHolderResult はポイント保有上位ユーザーの結果
type TopHolderResult struct {
	ID          string
	Username    string
	DisplayName string
	Balance     int64
}

// AnalyticsDataSource は分析用データソースインターフェース
type AnalyticsDataSource interface {
	// GetUserBalanceSummary はアクティブユーザーの残高サマリーを取得
	GetUserBalanceSummary(ctx context.Context) (*AnalyticsSummaryResult, error)

	// GetTopHolders はポイント保有上位ユーザーを取得
	GetTopHolders(ctx context.Context, limit int) ([]*TopHolderResult, error)

	// GetDailyStats は日別統計を取得
	GetDailyStats(ctx context.Context, since time.Time) ([]*DailyStatResult, error)

	// GetTransactionTypeBreakdown はトランザクション種別構成を取得
	GetTransactionTypeBreakdown(ctx context.Context) ([]*TypeBreakdownResult, error)

	// GetMonthlyIssuedPoints は今月の発行ポイント数を取得
	GetMonthlyIssuedPoints(ctx context.Context) (int64, error)

	// GetMonthlyTransactionCount は今月のトランザクション数を取得
	GetMonthlyTransactionCount(ctx context.Context) (int64, error)
}
