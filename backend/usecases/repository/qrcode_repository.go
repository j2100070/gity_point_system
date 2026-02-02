package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// QRCodeRepository はQRコードのリポジトリインターフェース
type QRCodeRepository interface {
	// Create は新しいQRコードを作成
	Create(qrCode *entities.QRCode) error

	// ReadByCode はコードでQRコードを検索
	ReadByCode(code string) (*entities.QRCode, error)

	// Read はIDでQRコードを検索
	Read(id uuid.UUID) (*entities.QRCode, error)

	// ReadListByUserID はユーザーのQRコード一覧を取得
	ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error)

	// Update はQRコードを更新
	Update(qrCode *entities.QRCode) error

	// DeleteExpired は期限切れQRコードを削除
	DeleteExpired() error
}
