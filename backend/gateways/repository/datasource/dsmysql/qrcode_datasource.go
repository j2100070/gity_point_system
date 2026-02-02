package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// QRCodeDataSource はMySQLのQRコードデータソースインターフェース
type QRCodeDataSource interface {
	// Insert は新しいQRコードを挿入
	Insert(qrCode *entities.QRCode) error

	// SelectByCode はコードでQRコードを検索
	SelectByCode(code string) (*entities.QRCode, error)

	// Select はIDでQRコードを検索
	Select(id uuid.UUID) (*entities.QRCode, error)

	// SelectListByUserID はユーザーのQRコード一覧を取得
	SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error)

	// Update はQRコードを更新
	Update(qrCode *entities.QRCode) error

	// DeleteExpired は期限切れQRコードを削除
	DeleteExpired() error
}
