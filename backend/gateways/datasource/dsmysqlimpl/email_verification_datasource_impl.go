package dsmysqlimpl

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailVerificationTokenModel はGORM用のメール認証トークンモデル
type EmailVerificationTokenModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     *uuid.UUID `gorm:"type:uuid"`
	Email      string     `gorm:"type:varchar(255);not null"`
	Token      string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	TokenType  string     `gorm:"type:varchar(50);not null"`
	ExpiresAt  time.Time  `gorm:"not null"`
	CreatedAt  time.Time  `gorm:"not null;default:now()"`
	VerifiedAt *time.Time
}

// TableName はテーブル名を指定
func (EmailVerificationTokenModel) TableName() string {
	return "email_verification_tokens"
}

// ToDomain はドメインモデルに変換
func (m *EmailVerificationTokenModel) ToDomain() *entities.EmailVerificationToken {
	return &entities.EmailVerificationToken{
		ID:         m.ID,
		UserID:     m.UserID,
		Email:      m.Email,
		Token:      m.Token,
		TokenType:  entities.TokenType(m.TokenType),
		ExpiresAt:  m.ExpiresAt,
		CreatedAt:  m.CreatedAt,
		VerifiedAt: m.VerifiedAt,
	}
}

// FromDomain はドメインモデルから変換
func (m *EmailVerificationTokenModel) FromDomain(token *entities.EmailVerificationToken) {
	m.ID = token.ID
	m.UserID = token.UserID
	m.Email = token.Email
	m.Token = token.Token
	m.TokenType = string(token.TokenType)
	m.ExpiresAt = token.ExpiresAt
	m.CreatedAt = token.CreatedAt
	m.VerifiedAt = token.VerifiedAt
}

// EmailVerificationDataSourceImpl はEmailVerificationDataSourceの実装
type EmailVerificationDataSourceImpl struct {
	db inframysql.DB
}

// NewEmailVerificationDataSource は新しいEmailVerificationDataSourceを作成
func NewEmailVerificationDataSource(db inframysql.DB) *EmailVerificationDataSourceImpl {
	return &EmailVerificationDataSourceImpl{db: db}
}

// Insert は新しいメール認証トークンを挿入
func (ds *EmailVerificationDataSourceImpl) Insert(ctx context.Context, token *entities.EmailVerificationToken) error {
	model := &EmailVerificationTokenModel{}
	model.FromDomain(token)

	return inframysql.GetDB(ctx, ds.db.GetDB()).Create(model).Error
}

// SelectByToken はトークンで検索
func (ds *EmailVerificationDataSourceImpl) SelectByToken(ctx context.Context, token string) (*entities.EmailVerificationToken, error) {
	var model EmailVerificationTokenModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はトークン情報を更新
func (ds *EmailVerificationDataSourceImpl) Update(ctx context.Context, token *entities.EmailVerificationToken) error {
	model := &EmailVerificationTokenModel{}
	model.FromDomain(token)

	return inframysql.GetDB(ctx, ds.db.GetDB()).Model(&EmailVerificationTokenModel{}).
		Where("id = ?", token.ID).
		Updates(map[string]interface{}{
			"verified_at": model.VerifiedAt,
		}).Error
}

// DeleteExpired は期限切れのトークンを削除
func (ds *EmailVerificationDataSourceImpl) DeleteExpired(ctx context.Context) error {
	return inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("expires_at < ?", time.Now()).
		Delete(&EmailVerificationTokenModel{}).Error
}

// DeleteByUserID はユーザーIDに紐づくトークンを削除
func (ds *EmailVerificationDataSourceImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("user_id = ?", userID).
		Delete(&EmailVerificationTokenModel{}).Error
}
