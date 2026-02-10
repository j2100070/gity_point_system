package entities

import (
	"time"

	"github.com/google/uuid"
)

// UsernameChangeHistory はユーザー名変更履歴
type UsernameChangeHistory struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	OldUsername string
	NewUsername string
	ChangedAt   time.Time
	ChangedBy   *uuid.UUID // 変更実行者（通常は本人）
	IPAddress   *string
}

// NewUsernameChangeHistory は新しいユーザー名変更履歴を作成
func NewUsernameChangeHistory(userID uuid.UUID, oldUsername, newUsername string, changedBy *uuid.UUID, ipAddress *string) *UsernameChangeHistory {
	return &UsernameChangeHistory{
		ID:          uuid.New(),
		UserID:      userID,
		OldUsername: oldUsername,
		NewUsername: newUsername,
		ChangedAt:   time.Now(),
		ChangedBy:   changedBy,
		IPAddress:   ipAddress,
	}
}
