package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserRole はユーザーの役割を表す型
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// AvatarType はアバタータイプを表す型
type AvatarType string

const (
	AvatarTypeGenerated AvatarType = "generated" // 自動生成アバター
	AvatarTypeUploaded  AvatarType = "uploaded"  // ユーザーアップロード
)

// User はユーザーエンティティ
type User struct {
	ID              uuid.UUID
	Username        string
	Email           string
	PasswordHash    string
	DisplayName     string
	FirstName       string // 名前（プロフィール表示用）
	LastName        string // 苗字（プロフィール表示用）
	Balance         int64  // ポイント残高
	Role            UserRole
	Version         int // 楽観的ロック用
	IsActive        bool
	AvatarURL       *string    // アバター画像URL
	AvatarType      AvatarType // アバタータイプ
	PersonalQRCode  string     // 個人固定QRコード（user:{user_id}形式）
	EmailVerified   bool       // メール認証済みか
	EmailVerifiedAt *time.Time // メール認証日時
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewUser は新しいユーザーを作成
func NewUser(username, email, passwordHash, displayName, firstName, lastName string) (*User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if passwordHash == "" {
		return nil, errors.New("password hash is required")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}

	userID := uuid.New()
	return &User{
		ID:             userID,
		Username:       username,
		Email:          email,
		PasswordHash:   passwordHash,
		DisplayName:    displayName,
		FirstName:      firstName,
		LastName:       lastName,
		Balance:        0,
		Role:           RoleUser,
		Version:        1,
		IsActive:       true,
		AvatarURL:      nil,
		AvatarType:     AvatarTypeGenerated,
		PersonalQRCode: GeneratePersonalQRCode(userID), // 個人QRコード生成
		EmailVerified:  false,                          // 初期は未認証
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// GeneratePersonalQRCode はユーザーIDから個人QRコードを生成
func GeneratePersonalQRCode(userID uuid.UUID) string {
	return "user:" + userID.String()
}

// IsAdmin はユーザーが管理者かどうかを確認
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanTransfer は送金可能かどうかを確認
func (u *User) CanTransfer(amount int64) error {
	if !u.IsActive {
		return errors.New("user is not active")
	}
	if u.Balance < amount {
		return errors.New("insufficient balance")
	}
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}

// Deduct はポイントを減算（楽観的ロック対応）
func (u *User) Deduct(amount int64) error {
	if err := u.CanTransfer(amount); err != nil {
		return err
	}
	u.Balance -= amount
	u.Version++
	u.UpdatedAt = time.Now()
	return nil
}

// Add はポイントを加算
func (u *User) Add(amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	u.Balance += amount
	u.Version++
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateRole はユーザーの役割を更新（管理者操作）
func (u *User) UpdateRole(newRole UserRole) error {
	if newRole != RoleUser && newRole != RoleAdmin {
		return errors.New("invalid role")
	}
	u.Role = newRole
	u.Version++
	u.UpdatedAt = time.Now()
	return nil
}

// Deactivate はユーザーを無効化
func (u *User) Deactivate() {
	u.IsActive = false
	u.Version++
	u.UpdatedAt = time.Now()
}

// Activate はユーザーを有効化
func (u *User) Activate() {
	u.IsActive = true
	u.Version++
	u.UpdatedAt = time.Now()
}

// UpdateProfile はプロフィール更新
func (u *User) UpdateProfile(displayName, email, firstName, lastName string) error {
	changed := false

	if displayName != "" && displayName != u.DisplayName {
		u.DisplayName = displayName
		changed = true
	}

	if email != "" && email != u.Email {
		u.Email = email
		u.EmailVerified = false // メール変更時は再認証必要
		u.EmailVerifiedAt = nil
		changed = true
	}

	if firstName != u.FirstName {
		u.FirstName = firstName
		changed = true
	}

	if lastName != u.LastName {
		u.LastName = lastName
		changed = true
	}

	if changed {
		u.Version++
		u.UpdatedAt = time.Now()
	}

	return nil
}

// UpdateUsername はユーザー名更新
func (u *User) UpdateUsername(newUsername string) error {
	if newUsername == "" {
		return errors.New("username is required")
	}
	if newUsername == u.Username {
		return errors.New("username is the same as current")
	}

	u.Username = newUsername
	u.Version++
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateAvatar はアバター更新
func (u *User) UpdateAvatar(avatarURL string, avatarType AvatarType) error {
	if avatarType != AvatarTypeGenerated && avatarType != AvatarTypeUploaded {
		return errors.New("invalid avatar type")
	}

	u.AvatarURL = &avatarURL
	u.AvatarType = avatarType
	u.Version++
	u.UpdatedAt = time.Now()
	return nil
}

// DeleteAvatar はアバター削除（自動生成に戻す）
func (u *User) DeleteAvatar() {
	u.AvatarURL = nil
	u.AvatarType = AvatarTypeGenerated
	u.Version++
	u.UpdatedAt = time.Now()
}

// VerifyEmail はメール認証
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	now := time.Now()
	u.EmailVerifiedAt = &now
	u.Version++
	u.UpdatedAt = now
}

// UpdatePassword はパスワード更新
func (u *User) UpdatePassword(newPasswordHash string) error {
	if newPasswordHash == "" {
		return errors.New("password hash is required")
	}

	u.PasswordHash = newPasswordHash
	u.Version++
	u.UpdatedAt = time.Now()
	return nil
}
