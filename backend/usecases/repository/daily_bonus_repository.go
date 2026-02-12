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

	// Update はデイリーボーナスを更新
	Update(ctx context.Context, bonus *entities.DailyBonus) error

	// Read はIDでデイリーボーナスを取得
	Read(ctx context.Context, id uuid.UUID) (*entities.DailyBonus, error)

	// ReadByUserAndDate はユーザーIDと日付でデイリーボーナスを取得
	ReadByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailyBonus, error)

	// ReadRecentByUser はユーザーの最近のデイリーボーナスを取得
	ReadRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error)

	// CountAllCompletedByUser はユーザーの全達成回数をカウント
	CountAllCompletedByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
