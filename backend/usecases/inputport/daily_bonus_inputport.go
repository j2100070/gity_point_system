package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// DailyBonusInputPort はデイリーボーナスのユースケースインターフェース
type DailyBonusInputPort interface {
	// GetTodayBonus は本日のボーナス状況を取得
	GetTodayBonus(ctx context.Context, req *GetTodayBonusRequest) (*GetTodayBonusResponse, error)

	// GetRecentBonuses は最近のボーナス履歴を取得
	GetRecentBonuses(ctx context.Context, req *GetRecentBonusesRequest) (*GetRecentBonusesResponse, error)

	// GetBonusSettings はボーナス設定を取得
	GetBonusSettings(ctx context.Context) (*BonusSettingsResponse, error)

	// UpdateBonusSettings はボーナス設定を更新（管理者用）
	UpdateBonusSettings(ctx context.Context, req *UpdateBonusSettingsRequest) error
}

// GetTodayBonusRequest は本日のボーナス状況取得リクエスト
type GetTodayBonusRequest struct {
	UserID uuid.UUID
}

// GetTodayBonusResponse は本日のボーナス状況取得レスポンス
type GetTodayBonusResponse struct {
	DailyBonus  *entities.DailyBonus // nil = 未獲得
	BonusPoints int64                // 設定されているボーナスポイント数
	TotalDays   int64                // これまでの獲得日数
}

// GetRecentBonusesRequest は最近のボーナス履歴取得リクエスト
type GetRecentBonusesRequest struct {
	UserID uuid.UUID
	Limit  int
}

// GetRecentBonusesResponse は最近のボーナス履歴取得レスポンス
type GetRecentBonusesResponse struct {
	Bonuses   []*entities.DailyBonus
	TotalDays int64
}

// BonusSettingsResponse はボーナス設定レスポンス
type BonusSettingsResponse struct {
	BonusPoints int64
}

// UpdateBonusSettingsRequest はボーナス設定更新リクエスト
type UpdateBonusSettingsRequest struct {
	BonusPoints int64
}
