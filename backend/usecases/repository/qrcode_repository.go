package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// QRCodeRepository はQRコードのリポジトリインターフェース
type QRCodeRepository interface {
	// Create は新しいQRコードを作成
	Create(ctx context.Context, qrCode *entities.QRCode) error

	// ReadByCode はコードでQRコードを検索
	ReadByCode(ctx context.Context, code string) (*entities.QRCode, error)

	// Read はIDでQRコードを検索
	Read(ctx context.Context, id uuid.UUID) (*entities.QRCode, error)

	// ReadListByUserID はユーザーのQRコード一覧を取得
	ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error)

	// Update はQRコードを更新
	Update(ctx context.Context, qrCode *entities.QRCode) error

	// DeleteExpired は期限切れQRコードを削除
	DeleteExpired(ctx context.Context) error
}
