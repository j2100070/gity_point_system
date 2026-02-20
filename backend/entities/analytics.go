package entities

import "time"

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
