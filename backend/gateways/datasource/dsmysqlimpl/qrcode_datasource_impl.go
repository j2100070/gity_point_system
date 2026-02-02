package dsmysqlimpl

import (
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QRCodeModel はGORM用のQRコードモデル
type QRCodeModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Code         string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Amount       *int64     `gorm:"type:bigint"`
	QRType       string     `gorm:"type:varchar(50);not null"`
	ExpiresAt    time.Time  `gorm:"not null;index"`
	UsedAt       *time.Time
	UsedByUserID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
}

// TableName はテーブル名を指定
func (QRCodeModel) TableName() string {
	return "qr_codes"
}

// ToDomain はドメインモデルに変換
func (q *QRCodeModel) ToDomain() *entities.QRCode {
	return &entities.QRCode{
		ID:           q.ID,
		UserID:       q.UserID,
		Code:         q.Code,
		Amount:       q.Amount,
		QRType:       entities.QRCodeType(q.QRType),
		ExpiresAt:    q.ExpiresAt,
		UsedAt:       q.UsedAt,
		UsedByUserID: q.UsedByUserID,
		CreatedAt:    q.CreatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (q *QRCodeModel) FromDomain(qrCode *entities.QRCode) {
	q.ID = qrCode.ID
	q.UserID = qrCode.UserID
	q.Code = qrCode.Code
	q.Amount = qrCode.Amount
	q.QRType = string(qrCode.QRType)
	q.ExpiresAt = qrCode.ExpiresAt
	q.UsedAt = qrCode.UsedAt
	q.UsedByUserID = qrCode.UsedByUserID
	q.CreatedAt = qrCode.CreatedAt
}

// QRCodeDataSourceImpl はQRCodeDataSourceの実装
type QRCodeDataSourceImpl struct {
	db inframysql.DB
}

// NewQRCodeDataSource は新しいQRCodeDataSourceを作成
func NewQRCodeDataSource(db inframysql.DB) dsmysql.QRCodeDataSource {
	return &QRCodeDataSourceImpl{db: db}
}

// Insert は新しいQRコードを挿入
func (ds *QRCodeDataSourceImpl) Insert(qrCode *entities.QRCode) error {
	model := &QRCodeModel{}
	model.FromDomain(qrCode)

	if err := ds.db.GetDB().Create(model).Error; err != nil {
		return err
	}

	*qrCode = *model.ToDomain()
	return nil
}

// SelectByCode はコードでQRコードを検索
func (ds *QRCodeDataSourceImpl) SelectByCode(code string) (*entities.QRCode, error) {
	var model QRCodeModel

	err := ds.db.GetDB().Where("code = ?", code).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("qr code not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Select はIDでQRコードを検索
func (ds *QRCodeDataSourceImpl) Select(id uuid.UUID) (*entities.QRCode, error) {
	var model QRCodeModel

	err := ds.db.GetDB().Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("qr code not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectListByUserID はユーザーのQRコード一覧を取得
func (ds *QRCodeDataSourceImpl) SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error) {
	var models []QRCodeModel

	err := ds.db.GetDB().
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	qrCodes := make([]*entities.QRCode, len(models))
	for i, model := range models {
		qrCodes[i] = model.ToDomain()
	}

	return qrCodes, nil
}

// Update はQRコードを更新
func (ds *QRCodeDataSourceImpl) Update(qrCode *entities.QRCode) error {
	model := &QRCodeModel{}
	model.FromDomain(qrCode)

	return ds.db.GetDB().Model(&QRCodeModel{}).
		Where("id = ?", qrCode.ID).
		Updates(map[string]interface{}{
			"used_at":        model.UsedAt,
			"used_by_user_id": model.UsedByUserID,
		}).Error
}

// DeleteExpired は期限切れQRコードを削除
func (ds *QRCodeDataSourceImpl) DeleteExpired() error {
	return ds.db.GetDB().
		Where("expires_at < ?", time.Now()).
		Delete(&QRCodeModel{}).Error
}
