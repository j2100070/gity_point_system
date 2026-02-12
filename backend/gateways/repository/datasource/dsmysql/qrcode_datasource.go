package dsmysql

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// QRCodeDataSource はMySQLのQRコードデータソースインターフェース
type QRCodeDataSource interface {
	// Insert は新しいQRコードを挿入
	Insert(ctx context.Context, qrCode *entities.QRCode) error

	// SelectByCode はコードでQRコードを検索
	SelectByCode(ctx context.Context, code string) (*entities.QRCode, error)

	// Select はIDでQRコードを検索
	Select(ctx context.Context, id uuid.UUID) (*entities.QRCode, error)

	// SelectListByUserID はユーザーのQRコード一覧を取得
	SelectListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error)

	// Update はQRコードを更新
	Update(ctx context.Context, qrCode *entities.QRCode) error

	// DeleteExpired は期限切れQRコードを削除
	DeleteExpired(ctx context.Context) error
}
