package dspostgresimpl

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionModel はGORM用のセッションモデル
type SessionModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	SessionToken string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	CSRFToken    string    `gorm:"type:varchar(255);not null"`
	IPAddress    string    `gorm:"type:varchar(100)"`
	UserAgent    string    `gorm:"type:text"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

// TableName はテーブル名を指定
func (SessionModel) TableName() string {
	return "sessions"
}

// ToDomain はドメインモデルに変換
func (s *SessionModel) ToDomain() *entities.Session {
	return &entities.Session{
		ID:           s.ID,
		UserID:       s.UserID,
		SessionToken: s.SessionToken,
		CSRFToken:    s.CSRFToken,
		IPAddress:    s.IPAddress,
		UserAgent:    s.UserAgent,
		ExpiresAt:    s.ExpiresAt,
		CreatedAt:    s.CreatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (s *SessionModel) FromDomain(session *entities.Session) {
	s.ID = session.ID
	s.UserID = session.UserID
	s.SessionToken = session.SessionToken
	s.CSRFToken = session.CSRFToken
	s.IPAddress = session.IPAddress
	s.UserAgent = session.UserAgent
	s.ExpiresAt = session.ExpiresAt
	s.CreatedAt = session.CreatedAt
}

// SessionDataSourceImpl はSessionDataSourceの実装
type SessionDataSourceImpl struct {
	db infrapostgres.DB
}

// NewSessionDataSource は新しいSessionDataSourceを作成
func NewSessionDataSource(db infrapostgres.DB) dsmysql.SessionDataSource {
	return &SessionDataSourceImpl{db: db}
}

// Insert は新しいセッションを挿入
func (ds *SessionDataSourceImpl) Insert(ctx context.Context, session *entities.Session) error {
	model := &SessionModel{}
	model.FromDomain(session)

	if err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Create(model).Error; err != nil {
		return err
	}

	*session = *model.ToDomain()
	return nil
}

// SelectByToken はトークンでセッションを検索
func (ds *SessionDataSourceImpl) SelectByToken(ctx context.Context, token string) (*entities.Session, error) {
	var model SessionModel

	err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("session_token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はセッションを更新
func (ds *SessionDataSourceImpl) Update(ctx context.Context, session *entities.Session) error {
	model := &SessionModel{}
	model.FromDomain(session)

	return infrapostgres.GetDB(ctx, ds.db.GetDB()).Model(&SessionModel{}).
		Where("id = ?", session.ID).
		Updates(map[string]interface{}{
			"expires_at": model.ExpiresAt,
		}).Error
}

// Delete はセッションを削除
func (ds *SessionDataSourceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("id = ?", id).Delete(&SessionModel{}).Error
}

// DeleteByUserID はユーザーの全セッションを削除
func (ds *SessionDataSourceImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("user_id = ?", userID).Delete(&SessionModel{}).Error
}

// DeleteExpired は期限切れセッションを削除
func (ds *SessionDataSourceImpl) DeleteExpired(ctx context.Context) error {
	return infrapostgres.GetDB(ctx, ds.db.GetDB()).
		Where("expires_at < ?", time.Now()).
		Delete(&SessionModel{}).Error
}
