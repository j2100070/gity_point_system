package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Session はセッションエンティティ
type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	SessionToken string
	CSRFToken    string
	IPAddress    string
	UserAgent    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// NewSession は新しいセッションを作成
func NewSession(userID uuid.UUID, ipAddress, userAgent string) (*Session, error) {
	sessionToken, err := GenerateSecureTokenBase64(32)
	if err != nil {
		return nil, err
	}

	csrfToken, err := GenerateSecureTokenBase64(32)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:           uuid.New(),
		UserID:       userID,
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24時間有効
		CreatedAt:    time.Now(),
	}, nil
}

// IsExpired はセッションが期限切れかどうかを確認
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// ValidateCSRF はCSRFトークンを検証
func (s *Session) ValidateCSRF(token string) error {
	if s.CSRFToken != token {
		return errors.New("invalid csrf token")
	}
	if s.IsExpired() {
		return errors.New("session expired")
	}
	return nil
}

// Refresh はセッションの有効期限を延長
func (s *Session) Refresh() {
	s.ExpiresAt = time.Now().Add(24 * time.Hour)
}

