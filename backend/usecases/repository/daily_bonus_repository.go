package repository

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// DailyBonusRepository はデイリーボーナスのリポジトリインターフェース
type DailyBonusRepository interface {
	// Create はデイリーボーナスを作成
	Create(ctx context.Context, bonus *entities.DailyBonus) error

	// ReadByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
	ReadByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailyBonus, error)

	// ReadRecentByUser はユーザーの最近のデイリーボーナスを取得
	ReadRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error)

	// CountByUser はユーザーのボーナス獲得日数をカウント
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)

	// GetLastPolledAt は前回ポーリング時刻を取得
	GetLastPolledAt(ctx context.Context) (time.Time, error)

	// UpdateLastPolledAt はポーリング時刻を更新
	UpdateLastPolledAt(ctx context.Context, t time.Time) error

	// MarkAsViewed はデイリーボーナスを閲覧済みにする
	MarkAsViewed(ctx context.Context, id uuid.UUID) error

	// UpdateDrawnResult は抽選結果を更新する（ルーレット実行時）
	UpdateDrawnResult(ctx context.Context, id uuid.UUID, bonusPoints int64, lotteryTierID *uuid.UUID, lotteryTierName string) error
}
