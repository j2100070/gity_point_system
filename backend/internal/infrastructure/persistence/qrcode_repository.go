package persistence

import (
	"errors"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QRCodeModel はGORM用のQRコードモデル
type QRCodeModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Code         string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Amount       *int64     `gorm:"check:amount > 0"`
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
func (q *QRCodeModel) ToDomain() *domain.QRCode {
	return &domain.QRCode{
		ID:           q.ID,
		UserID:       q.UserID,
		Code:         q.Code,
		Amount:       q.Amount,
		QRType:       domain.QRCodeType(q.QRType),
		ExpiresAt:    q.ExpiresAt,
		UsedAt:       q.UsedAt,
		UsedByUserID: q.UsedByUserID,
		CreatedAt:    q.CreatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (q *QRCodeModel) FromDomain(qrCode *domain.QRCode) {
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

// QRCodeRepositoryImpl はQRCodeRepositoryの実装
type QRCodeRepositoryImpl struct {
	db *gorm.DB
}

// NewQRCodeRepository は新しいQRCodeRepositoryを作成
func NewQRCodeRepository(db *gorm.DB) domain.QRCodeRepository {
	return &QRCodeRepositoryImpl{db: db}
}

// Create は新しいQRコードを作成
func (r *QRCodeRepositoryImpl) Create(qrCode *domain.QRCode) error {
	model := &QRCodeModel{}
	model.FromDomain(qrCode)

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	*qrCode = *model.ToDomain()
	return nil
}

// FindByCode はコードでQRコードを検索
func (r *QRCodeRepositoryImpl) FindByCode(code string) (*domain.QRCode, error) {
	var model QRCodeModel

	err := r.db.Where("code = ?", code).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("qr code not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// FindByID はIDでQRコードを検索
func (r *QRCodeRepositoryImpl) FindByID(id uuid.UUID) (*domain.QRCode, error) {
	var model QRCodeModel

	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("qr code not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// ListByUserID はユーザーのQRコード一覧を取得
func (r *QRCodeRepositoryImpl) ListByUserID(userID uuid.UUID, offset, limit int) ([]*domain.QRCode, error) {
	var models []QRCodeModel

	err := r.db.Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	qrCodes := make([]*domain.QRCode, len(models))
	for i, model := range models {
		qrCodes[i] = model.ToDomain()
	}

	return qrCodes, nil
}

// Update はQRコードを更新
func (r *QRCodeRepositoryImpl) Update(qrCode *domain.QRCode) error {
	model := &QRCodeModel{}
	model.FromDomain(qrCode)

	return r.db.Save(model).Error
}

// DeleteExpired は期限切れQRコードを削除
func (r *QRCodeRepositoryImpl) DeleteExpired() error {
	return r.db.Where("expires_at < ? AND used_at IS NULL", time.Now()).Delete(&QRCodeModel{}).Error
}
