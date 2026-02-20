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

	// GetBonusSettings はボーナス設定を取得（管理者用）
	GetBonusSettings(ctx context.Context) (*BonusSettingsResponse, error)

	// UpdateLotteryTiers は抽選ティアを一括更新（管理者用）
	UpdateLotteryTiers(ctx context.Context, req *UpdateLotteryTiersRequest) error

	// MarkBonusViewed はボーナスを閲覧済みにする
	MarkBonusViewed(ctx context.Context, req *MarkBonusViewedRequest) error

	// DrawLotteryAndGrant はルーレットを実行しポイントを付与する
	DrawLotteryAndGrant(ctx context.Context, req *DrawLotteryRequest) (*DrawLotteryResponse, error)
}

// GetTodayBonusRequest は本日のボーナス状況取得リクエスト
type GetTodayBonusRequest struct {
	UserID uuid.UUID
}

// GetTodayBonusResponse は本日のボーナス状況取得レスポンス
type GetTodayBonusResponse struct {
	DailyBonus       *entities.DailyBonus // nil = 未獲得
	BonusPoints      int64                // 設定されているボーナスポイント数（フォールバック用）
	TotalDays        int64                // これまでの獲得日数
	IsLotteryPending bool                 // くじ引きアニメーション未再生
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
	BonusPoints  int64                   // フォールバック固定ポイント
	LotteryTiers []*entities.LotteryTier // 抽選ティア一覧
}

// LotteryTierInput は抽選ティアの入力
type LotteryTierInput struct {
	Name         string
	Points       int64
	Probability  float64
	DisplayOrder int
}

// UpdateLotteryTiersRequest は抽選ティア一括更新リクエスト
type UpdateLotteryTiersRequest struct {
	Tiers []LotteryTierInput
}

// MarkBonusViewedRequest はボーナス閲覧済みリクエスト
type MarkBonusViewedRequest struct {
	BonusID uuid.UUID
	UserID  uuid.UUID
}

// DrawLotteryRequest はルーレット実行リクエスト
type DrawLotteryRequest struct {
	UserID uuid.UUID
}

// DrawLotteryResponse はルーレット実行レスポンス
type DrawLotteryResponse struct {
	BonusPoints     int64
	LotteryTierName string
	BonusID         uuid.UUID
}
