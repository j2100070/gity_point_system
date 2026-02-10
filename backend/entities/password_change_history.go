package entities

import (
	"time"

	"github.com/google/uuid"
)

// PasswordChangeHistory はパスワード変更履歴（セキュリティ監査用）
type PasswordChangeHistory struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ChangedAt time.Time
	IPAddress *string
	UserAgent *string
}

// NewPasswordChangeHistory は新しいパスワード変更履歴を作成
func NewPasswordChangeHistory(userID uuid.UUID, ipAddress, userAgent *string) *PasswordChangeHistory {
	return &PasswordChangeHistory{
		ID:        uuid.New(),
		UserID:    userID,
		ChangedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}
