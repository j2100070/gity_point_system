package entities

import (
	"time"

	"github.com/google/uuid"
)

// PointBatchSourceType はポイントバッチのソースタイプ
type PointBatchSourceType string

const (
	PointBatchSourceTransfer    PointBatchSourceType = "transfer"
	PointBatchSourceAdminGrant  PointBatchSourceType = "admin_grant"
	PointBatchSourceDailyBonus  PointBatchSourceType = "daily_bonus"
	PointBatchSourceSystemGrant PointBatchSourceType = "system_grant"
	PointBatchSourceMigration   PointBatchSourceType = "migration"
)

// PointExpirationDuration はポイントの有効期限（3ヶ月）
const PointExpirationMonths = 3

// PointBatch はポイントバッチエンティティ
// 獲得ポイントをバッチ単位で追跡し、FIFO消費と期限切れ処理を行う
type PointBatch struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	OriginalAmount      int64
	RemainingAmount     int64
	SourceType          PointBatchSourceType
	SourceTransactionID *uuid.UUID
	ExpiresAt           time.Time
	CreatedAt           time.Time
}

// NewPointBatch は新しいポイントバッチを作成
func NewPointBatch(userID uuid.UUID, amount int64, sourceType PointBatchSourceType, txID *uuid.UUID, now time.Time) *PointBatch {
	return &PointBatch{
		ID:                  uuid.New(),
		UserID:              userID,
		OriginalAmount:      amount,
		RemainingAmount:     amount,
		SourceType:          sourceType,
		SourceTransactionID: txID,
		ExpiresAt:           now.AddDate(0, PointExpirationMonths, 0),
		CreatedAt:           now,
	}
}
