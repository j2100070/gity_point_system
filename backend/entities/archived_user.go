package entities

import (
	"time"

	"github.com/google/uuid"
)

// ArchivedUser はアーカイブされた（削除された）ユーザー
type ArchivedUser struct {
	ID                uuid.UUID
	Username          string
	Email             string
	PasswordHash      string
	DisplayName       string
	Balance           int64
	Role              UserRole
	AvatarURL         *string
	AvatarType        AvatarType
	EmailVerified     bool
	EmailVerifiedAt   *time.Time
	ArchivedAt        time.Time
	ArchivedBy        *uuid.UUID // 削除実行者（自分 or 管理者）
	DeletionReason    *string    // 削除理由
	OriginalCreatedAt time.Time
	OriginalUpdatedAt time.Time
}

// ToArchivedUser はUserをArchivedUserに変換
func (u *User) ToArchivedUser(archivedBy *uuid.UUID, reason *string) *ArchivedUser {
	return &ArchivedUser{
		ID:                u.ID,
		Username:          u.Username,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		DisplayName:       u.DisplayName,
		Balance:           u.Balance,
		Role:              u.Role,
		AvatarURL:         u.AvatarURL,
		AvatarType:        u.AvatarType,
		EmailVerified:     u.EmailVerified,
		EmailVerifiedAt:   u.EmailVerifiedAt,
		ArchivedAt:        time.Now(),
		ArchivedBy:        archivedBy,
		DeletionReason:    reason,
		OriginalCreatedAt: u.CreatedAt,
		OriginalUpdatedAt: u.UpdatedAt,
	}
}

// RestoreToUser はArchivedUserをUserに復元
func (au *ArchivedUser) RestoreToUser() *User {
	return &User{
		ID:              au.ID,
		Username:        au.Username,
		Email:           au.Email,
		PasswordHash:    au.PasswordHash,
		DisplayName:     au.DisplayName,
		Balance:         au.Balance,
		Role:            au.Role,
		AvatarURL:       au.AvatarURL,
		AvatarType:      au.AvatarType,
		EmailVerified:   au.EmailVerified,
		EmailVerifiedAt: au.EmailVerifiedAt,
		Version:         1, // 復元時はバージョンをリセット
		IsActive:        true,
		CreatedAt:       au.OriginalCreatedAt,
		UpdatedAt:       time.Now(),
	}
}
