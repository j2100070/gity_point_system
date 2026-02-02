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

// User はユーザーエンティティ
type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	DisplayName  string
	Balance      int64      // ポイント残高
	Role         UserRole
	Version      int        // 楽観的ロック用
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time // ソフトデリート
}

// NewUser は新しいユーザーを作成
func NewUser(username, email, passwordHash, displayName string) (*User, error) {
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

	return &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Balance:      0,
		Role:         RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
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
