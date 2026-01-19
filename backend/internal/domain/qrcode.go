package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
)

// QRCodeType はQRコードのタイプ
type QRCodeType string

const (
	QRCodeTypeReceive QRCodeType = "receive" // ポイント受取用
	QRCodeTypeSend    QRCodeType = "send"    // ポイント送信用
)

// QRCode はQRコードエンティティ
type QRCode struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Code        string      // ランダム生成コード
	Amount      *int64      // nil=送信者が金額指定、値あり=固定額
	QRType      QRCodeType
	ExpiresAt   time.Time
	UsedAt      *time.Time
	UsedByUserID *uuid.UUID // 使用したユーザー
	CreatedAt   time.Time
}

// NewReceiveQRCode はポイント受取用QRコードを作成
func NewReceiveQRCode(userID uuid.UUID, amount *int64) (*QRCode, error) {
	if amount != nil && *amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	code, err := generateQRCode()
	if err != nil {
		return nil, err
	}

	return &QRCode{
		ID:        uuid.New(),
		UserID:    userID,
		Code:      code,
		Amount:    amount,
		QRType:    QRCodeTypeReceive,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5分間有効
		CreatedAt: time.Now(),
	}, nil
}

// NewSendQRCode はポイント送信用QRコードを作成
func NewSendQRCode(userID uuid.UUID, amount int64) (*QRCode, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	code, err := generateQRCode()
	if err != nil {
		return nil, err
	}

	return &QRCode{
		ID:        uuid.New(),
		UserID:    userID,
		Code:      code,
		Amount:    &amount,
		QRType:    QRCodeTypeSend,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5分間有効
		CreatedAt: time.Now(),
	}, nil
}

// IsExpired はQRコードが期限切れかどうかを確認
func (q *QRCode) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}

// IsUsed はQRコードが使用済みかどうかを確認
func (q *QRCode) IsUsed() bool {
	return q.UsedAt != nil
}

// MarkAsUsed はQRコードを使用済みにする
func (q *QRCode) MarkAsUsed(userID uuid.UUID) error {
	if q.IsUsed() {
		return errors.New("qr code already used")
	}
	if q.IsExpired() {
		return errors.New("qr code expired")
	}
	now := time.Now()
	q.UsedAt = &now
	q.UsedByUserID = &userID
	return nil
}

// CanBeUsedBy はQRコードが使用可能かどうかを確認
func (q *QRCode) CanBeUsedBy(userID uuid.UUID) error {
	if q.IsExpired() {
		return errors.New("qr code expired")
	}
	if q.IsUsed() {
		return errors.New("qr code already used")
	}
	if q.UserID == userID {
		return errors.New("cannot use your own qr code")
	}
	return nil
}

// QRCodeRepository はQRコードのリポジトリインターフェース
type QRCodeRepository interface {
	// Create は新しいQRコードを作成
	Create(qrCode *QRCode) error

	// FindByCode はコードでQRコードを検索
	FindByCode(code string) (*QRCode, error)

	// FindByID はIDでQRコードを検索
	FindByID(id uuid.UUID) (*QRCode, error)

	// ListByUserID はユーザーのQRコード一覧を取得
	ListByUserID(userID uuid.UUID, offset, limit int) ([]*QRCode, error)

	// Update はQRコードを更新
	Update(qrCode *QRCode) error

	// DeleteExpired は期限切れQRコードを削除
	DeleteExpired() error
}

// generateQRCode は安全なランダムQRコードを生成
func generateQRCode() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
