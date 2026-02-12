package entities

import (
	"time"

	"github.com/google/uuid"
)

// DailyBonus はデイリーボーナスの達成状況を表すエンティティ
type DailyBonus struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	BonusDate time.Time // JSTの日付

	// 各ボーナスの達成状況
	LoginCompleted      bool
	LoginCompletedAt    *time.Time
	TransferCompleted   bool
	TransferCompletedAt *time.Time
	TransferTxID        *uuid.UUID
	ExchangeCompleted   bool
	ExchangeCompletedAt *time.Time
	ExchangeID          *uuid.UUID

	// 全て達成
	AllCompleted   bool
	AllCompletedAt *time.Time

	// 付与されたボーナスポイント
	TotalBonusPoints int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ボーナスポイント定数
const (
	LoginBonusPoints    int64 = 10  // ログインボーナス
	TransferBonusPoints int64 = 10  // 送金ボーナス
	ExchangeBonusPoints int64 = 10  // 交換ボーナス
	AllCompleteBonusPoints int64 = 50  // 全達成ボーナス（追加分は20p = 50 - 30）
)

// NewDailyBonus は新しいDailyBonusを作成
func NewDailyBonus(userID uuid.UUID, bonusDate time.Time) *DailyBonus {
	now := time.Now()
	return &DailyBonus{
		ID:               uuid.New(),
		UserID:           userID,
		BonusDate:        bonusDate,
		LoginCompleted:   false,
		TransferCompleted: false,
		ExchangeCompleted: false,
		AllCompleted:     false,
		TotalBonusPoints: 0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// CompleteLogin はログインボーナスを達成済みにする
func (db *DailyBonus) CompleteLogin() int64 {
	if db.LoginCompleted {
		return 0 // 既に達成済み
	}

	now := time.Now()
	db.LoginCompleted = true
	db.LoginCompletedAt = &now
	db.TotalBonusPoints += LoginBonusPoints
	db.UpdatedAt = now

	// 全達成チェック
	db.checkAllCompleted()

	return LoginBonusPoints
}

// CompleteTransfer は送金ボーナスを達成済みにする
func (db *DailyBonus) CompleteTransfer(transactionID uuid.UUID) int64 {
	if db.TransferCompleted {
		return 0 // 既に達成済み
	}

	now := time.Now()
	db.TransferCompleted = true
	db.TransferCompletedAt = &now
	db.TransferTxID = &transactionID
	db.TotalBonusPoints += TransferBonusPoints
	db.UpdatedAt = now

	// 全達成チェック
	db.checkAllCompleted()

	return TransferBonusPoints
}

// CompleteExchange は商品交換ボーナスを達成済みにする
func (db *DailyBonus) CompleteExchange(exchangeID uuid.UUID) int64 {
	if db.ExchangeCompleted {
		return 0 // 既に達成済み
	}

	now := time.Now()
	db.ExchangeCompleted = true
	db.ExchangeCompletedAt = &now
	db.ExchangeID = &exchangeID
	db.TotalBonusPoints += ExchangeBonusPoints
	db.UpdatedAt = now

	// 全達成チェック
	db.checkAllCompleted()

	return ExchangeBonusPoints
}

// checkAllCompleted は全てのボーナスが達成されているかチェックし、追加ボーナスを付与
func (db *DailyBonus) checkAllCompleted() {
	if db.AllCompleted {
		return // 既に全達成済み
	}

	if db.LoginCompleted && db.TransferCompleted && db.ExchangeCompleted {
		now := time.Now()
		db.AllCompleted = true
		db.AllCompletedAt = &now
		// 追加ボーナス: 50 - (10 + 10 + 10) = 20ポイント
		additionalBonus := AllCompleteBonusPoints - (LoginBonusPoints + TransferBonusPoints + ExchangeBonusPoints)
		db.TotalBonusPoints += additionalBonus
		db.UpdatedAt = now
	}
}

// GetCompletedCount は達成済みボーナスの数を返す
func (db *DailyBonus) GetCompletedCount() int {
	count := 0
	if db.LoginCompleted {
		count++
	}
	if db.TransferCompleted {
		count++
	}
	if db.ExchangeCompleted {
		count++
	}
	return count
}

// GetRemainingBonus は未達成のボーナスで獲得できる残りポイントを返す
func (db *DailyBonus) GetRemainingBonus() int64 {
	maxBonus := AllCompleteBonusPoints
	return maxBonus - db.TotalBonusPoints
}
