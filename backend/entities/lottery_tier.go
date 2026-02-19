package entities

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// LotteryTier はボーナス抽選ティア
type LotteryTier struct {
	ID           uuid.UUID
	Name         string
	Points       int64
	Probability  float64 // 確率（%）例: 1.00 = 1%
	DisplayOrder int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewLotteryTier は新しいLotteryTierを作成
func NewLotteryTier(name string, points int64, probability float64, displayOrder int) *LotteryTier {
	now := time.Now()
	return &LotteryTier{
		ID:           uuid.New(),
		Name:         name,
		Points:       points,
		Probability:  probability,
		DisplayOrder: displayOrder,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// DrawLottery は確率に基づいて抽選を実行する
// ティアが空またはすべて0%の場合はnilを返す（ボーナスなし）
// 合計確率が100%未満の場合、残りの確率は「ボーナスなし」扱い
func DrawLottery(tiers []*LotteryTier) *LotteryTier {
	if len(tiers) == 0 {
		return nil
	}

	// 0〜100のランダム値を生成
	roll := rand.Float64() * 100.0

	cumulative := 0.0
	for _, tier := range tiers {
		cumulative += tier.Probability
		if roll < cumulative {
			return tier
		}
	}

	// 合計確率が100%未満の場合、ここに到達 → ボーナスなし
	return nil
}
