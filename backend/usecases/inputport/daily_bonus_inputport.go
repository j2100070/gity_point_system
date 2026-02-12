package inputport

import (
	"context"

	"time"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// DailyBonusInputPort はデイリーボーナスのユースケースインターフェース
type DailyBonusInputPort interface {
	// CheckLoginBonus はログインボーナスをチェックして付与
	CheckLoginBonus(ctx context.Context, req *CheckLoginBonusRequest) (*CheckLoginBonusResponse, error)

	// CheckTransferBonus は送金ボーナスをチェックして付与
	CheckTransferBonus(ctx context.Context, req *CheckTransferBonusRequest) (*CheckTransferBonusResponse, error)

	// CheckExchangeBonus は交換ボーナスをチェックして付与
	CheckExchangeBonus(ctx context.Context, req *CheckExchangeBonusRequest) (*CheckExchangeBonusResponse, error)

	// GetTodayBonus は本日のボーナス状況を取得
	GetTodayBonus(ctx context.Context, req *GetTodayBonusRequest) (*GetTodayBonusResponse, error)

	// GetRecentBonuses は最近のボーナス履歴を取得
	GetRecentBonuses(ctx context.Context, req *GetRecentBonusesRequest) (*GetRecentBonusesResponse, error)
}

// CheckLoginBonusRequest はログインボーナスチェックリクエスト
type CheckLoginBonusRequest struct {
	UserID uuid.UUID
	Date   time.Time // JSTの日付
}

// CheckLoginBonusResponse はログインボーナスチェックレスポンス
type CheckLoginBonusResponse struct {
	DailyBonus   *entities.DailyBonus
	BonusAwarded int64 // 付与されたポイント（0の場合は既に達成済み）
	User         *entities.User
}

// CheckTransferBonusRequest は送金ボーナスチェックリクエスト
type CheckTransferBonusRequest struct {
	UserID        uuid.UUID
	TransactionID uuid.UUID
	Date          time.Time // JSTの日付
}

// CheckTransferBonusResponse は送金ボーナスチェックレスポンス
type CheckTransferBonusResponse struct {
	DailyBonus   *entities.DailyBonus
	BonusAwarded int64 // 付与されたポイント
	User         *entities.User
}

// CheckExchangeBonusRequest は交換ボーナスチェックリクエスト
type CheckExchangeBonusRequest struct {
	UserID     uuid.UUID
	ExchangeID uuid.UUID
	Date       time.Time // JSTの日付
}

// CheckExchangeBonusResponse は交換ボーナスチェックレスポンス
type CheckExchangeBonusResponse struct {
	DailyBonus   *entities.DailyBonus
	BonusAwarded int64 // 付与されたポイント
	User         *entities.User
}

// GetTodayBonusRequest は本日のボーナス状況取得リクエスト
type GetTodayBonusRequest struct {
	UserID uuid.UUID
}

// GetTodayBonusResponse は本日のボーナス状況取得レスポンス
type GetTodayBonusResponse struct {
	DailyBonus           *entities.DailyBonus
	AllCompletedCount    int64 // これまでの全達成回数
	CanClaimLoginBonus   bool
	CanClaimTransferBonus bool
	CanClaimExchangeBonus bool
}

// GetRecentBonusesRequest は最近のボーナス履歴取得リクエスト
type GetRecentBonusesRequest struct {
	UserID uuid.UUID
	Limit  int
}

// GetRecentBonusesResponse は最近のボーナス履歴取得レスポンス
type GetRecentBonusesResponse struct {
	Bonuses           []*entities.DailyBonus
	AllCompletedCount int64
}
