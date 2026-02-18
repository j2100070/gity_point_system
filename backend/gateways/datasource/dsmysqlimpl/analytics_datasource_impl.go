package dsmysqlimpl

import (
	"context"
	"time"

	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
)

// AnalyticsDataSourceImpl は分析用データソースの実装
type AnalyticsDataSourceImpl struct {
	db inframysql.DB
}

// NewAnalyticsDataSource は新しいAnalyticsDataSourceを作成
func NewAnalyticsDataSource(db inframysql.DB) dsmysql.AnalyticsDataSource {
	return &AnalyticsDataSourceImpl{db: db}
}

// GetUserBalanceSummary はアクティブユーザーの残高サマリーを取得
func (ds *AnalyticsDataSourceImpl) GetUserBalanceSummary(ctx context.Context) (*dsmysql.AnalyticsSummaryResult, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	var result struct {
		TotalBalance   int64
		AverageBalance float64
		ActiveUsers    int64
	}

	err := db.Table("users").
		Select("COALESCE(SUM(balance), 0) as total_balance, COALESCE(AVG(balance), 0) as average_balance, COUNT(*) as active_users").
		Where("is_active = ?", true).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &dsmysql.AnalyticsSummaryResult{
		TotalBalance:   result.TotalBalance,
		AverageBalance: result.AverageBalance,
		ActiveUsers:    result.ActiveUsers,
	}, nil
}

// GetTopHolders はポイント保有上位ユーザーを取得
func (ds *AnalyticsDataSourceImpl) GetTopHolders(ctx context.Context, limit int) ([]*dsmysql.TopHolderResult, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	var results []struct {
		ID          string
		Username    string
		DisplayName string `gorm:"column:display_name"`
		Balance     int64
	}

	err := db.Table("users").
		Select("id, username, display_name, balance").
		Where("is_active = ?", true).
		Order("balance DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	holders := make([]*dsmysql.TopHolderResult, 0, len(results))
	for _, r := range results {
		holders = append(holders, &dsmysql.TopHolderResult{
			ID:          r.ID,
			Username:    r.Username,
			DisplayName: r.DisplayName,
			Balance:     r.Balance,
		})
	}
	return holders, nil
}

// GetDailyStats は日別統計を取得
func (ds *AnalyticsDataSourceImpl) GetDailyStats(ctx context.Context, since time.Time) ([]*dsmysql.DailyStatResult, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	var results []struct {
		Date        time.Time
		Issued      int64
		Consumed    int64
		Transferred int64
	}

	err := db.Table("transactions").
		Select(`
			DATE(created_at) as date,
			COALESCE(SUM(CASE WHEN transaction_type IN ('admin_grant', 'system_grant') THEN amount ELSE 0 END), 0) as issued,
			COALESCE(SUM(CASE WHEN transaction_type IN ('admin_deduct', 'system_expire') THEN amount ELSE 0 END), 0) as consumed,
			COALESCE(SUM(CASE WHEN transaction_type = 'transfer' THEN amount ELSE 0 END), 0) as transferred
		`).
		Where("created_at >= ? AND status = ?", since, "completed").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make([]*dsmysql.DailyStatResult, 0, len(results))
	for _, r := range results {
		stats = append(stats, &dsmysql.DailyStatResult{
			Date:        r.Date,
			Issued:      r.Issued,
			Consumed:    r.Consumed,
			Transferred: r.Transferred,
		})
	}
	return stats, nil
}

// GetTransactionTypeBreakdown はトランザクション種別構成を取得
func (ds *AnalyticsDataSourceImpl) GetTransactionTypeBreakdown(ctx context.Context) ([]*dsmysql.TypeBreakdownResult, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	var results []struct {
		Type        string `gorm:"column:transaction_type"`
		Count       int64
		TotalAmount int64 `gorm:"column:total_amount"`
	}

	err := db.Table("transactions").
		Select("transaction_type, COUNT(*) as count, COALESCE(SUM(amount), 0) as total_amount").
		Where("status = ?", "completed").
		Group("transaction_type").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	breakdowns := make([]*dsmysql.TypeBreakdownResult, 0, len(results))
	for _, r := range results {
		breakdowns = append(breakdowns, &dsmysql.TypeBreakdownResult{
			Type:        r.Type,
			Count:       r.Count,
			TotalAmount: r.TotalAmount,
		})
	}
	return breakdowns, nil
}

// GetMonthlyIssuedPoints は今月の発行ポイント数を取得
func (ds *AnalyticsDataSourceImpl) GetMonthlyIssuedPoints(ctx context.Context) (int64, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var result struct {
		Total int64
	}

	err := db.Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("transaction_type IN (?, ?) AND status = ? AND created_at >= ?",
			"admin_grant", "system_grant", "completed", monthStart).
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// GetMonthlyTransactionCount は今月のトランザクション数を取得
func (ds *AnalyticsDataSourceImpl) GetMonthlyTransactionCount(ctx context.Context) (int64, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := db.Table("transactions").
		Where("created_at >= ?", monthStart).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
