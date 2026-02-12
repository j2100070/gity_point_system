package qrcode

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// RepositoryImpl はQRCodeRepositoryの実装
type RepositoryImpl struct {
	qrcodeDS dsmysql.QRCodeDataSource
	logger   entities.Logger
}

// NewQRCodeRepository は新しいQRCodeRepositoryを作成
func NewQRCodeRepository(
	qrcodeDS dsmysql.QRCodeDataSource,
	logger entities.Logger,
) repository.QRCodeRepository {
	return &RepositoryImpl{
		qrcodeDS: qrcodeDS,
		logger:   logger,
	}
}

// Create は新しいQRコードを作成
func (r *RepositoryImpl) Create(ctx context.Context, qrCode *entities.QRCode) error {
	r.logger.Debug("Creating QR code", entities.NewField("user_id", qrCode.UserID))
	return r.qrcodeDS.Insert(ctx, qrCode)
}

// ReadByCode はコードでQRコードを検索
func (r *RepositoryImpl) ReadByCode(ctx context.Context, code string) (*entities.QRCode, error) {
	return r.qrcodeDS.SelectByCode(ctx, code)
}

// Read はIDでQRコードを検索
func (r *RepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.QRCode, error) {
	return r.qrcodeDS.Select(ctx, id)
}

// ReadListByUserID はユーザーのQRコード一覧を取得
func (r *RepositoryImpl) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error) {
	return r.qrcodeDS.SelectListByUserID(ctx, userID, offset, limit)
}

// Update はQRコードを更新
func (r *RepositoryImpl) Update(ctx context.Context, qrCode *entities.QRCode) error {
	r.logger.Debug("Updating QR code", entities.NewField("qrcode_id", qrCode.ID))
	return r.qrcodeDS.Update(ctx, qrCode)
}

// DeleteExpired は期限切れQRコードを削除
func (r *RepositoryImpl) DeleteExpired(ctx context.Context) error {
	r.logger.Debug("Deleting expired QR codes")
	return r.qrcodeDS.DeleteExpired(ctx)
}
