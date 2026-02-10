package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TokenType はトークンタイプ
type TokenType string

const (
	TokenTypeRegistration TokenType = "registration"  // 新規登録時のメール認証
	TokenTypeEmailChange  TokenType = "email_change" // メールアドレス変更時の認証
)

// EmailVerificationToken はメール認証トークン
type EmailVerificationToken struct {
	ID         uuid.UUID
	UserID     *uuid.UUID // 登録時はnull
	Email      string
	Token      string
	TokenType  TokenType
	ExpiresAt  time.Time
	CreatedAt  time.Time
	VerifiedAt *time.Time
}

// NewEmailVerificationToken は新しいメール認証トークンを作成
func NewEmailVerificationToken(userID *uuid.UUID, email string, tokenType TokenType) (*EmailVerificationToken, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if tokenType != TokenTypeRegistration && tokenType != TokenTypeEmailChange {
		return nil, errors.New("invalid token type")
	}

	// ランダムトークン生成（32バイト = 64文字の16進数）
	token, err := GenerateSecureTokenHex(32)
	if err != nil {
		return nil, err
	}

	return &EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		TokenType: tokenType,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24時間有効
		CreatedAt: time.Now(),
	}, nil
}

// IsExpired はトークンが期限切れかどうかを確認
func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsVerified はトークンが既に検証済みかどうかを確認
func (t *EmailVerificationToken) IsVerified() bool {
	return t.VerifiedAt != nil
}

// Verify はトークンを検証済みにする
func (t *EmailVerificationToken) Verify() error {
	if t.IsVerified() {
		return errors.New("token already verified")
	}
	if t.IsExpired() {
		return errors.New("token expired")
	}

	now := time.Now()
	t.VerifiedAt = &now
	return nil
}

