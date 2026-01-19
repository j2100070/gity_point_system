package persistence

import (
	"errors"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionModel はGORM用のセッションモデル
type SessionModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	SessionToken string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	CSRFToken    string    `gorm:"type:varchar(255);not null"`
	IPAddress    string    `gorm:"type:inet"`
	UserAgent    string    `gorm:"type:text"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

// TableName はテーブル名を指定
func (SessionModel) TableName() string {
	return "sessions"
}

// ToDomain はドメインモデルに変換
func (s *SessionModel) ToDomain() *domain.Session {
	return &domain.Session{
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
func (s *SessionModel) FromDomain(session *domain.Session) {
	s.ID = session.ID
	s.UserID = session.UserID
	s.SessionToken = session.SessionToken
	s.CSRFToken = session.CSRFToken
	s.IPAddress = session.IPAddress
	s.UserAgent = session.UserAgent
	s.ExpiresAt = session.ExpiresAt
	s.CreatedAt = session.CreatedAt
}

// SessionRepositoryImpl はSessionRepositoryの実装
type SessionRepositoryImpl struct {
	db *gorm.DB
}

// NewSessionRepository は新しいSessionRepositoryを作成
func NewSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &SessionRepositoryImpl{db: db}
}

// Create は新しいセッションを作成
func (r *SessionRepositoryImpl) Create(session *domain.Session) error {
	model := &SessionModel{}
	model.FromDomain(session)

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	*session = *model.ToDomain()
	return nil
}

// FindByToken はトークンでセッションを検索
func (r *SessionRepositoryImpl) FindByToken(token string) (*domain.Session, error) {
	var model SessionModel

	err := r.db.Where("session_token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はセッションを更新
func (r *SessionRepositoryImpl) Update(session *domain.Session) error {
	model := &SessionModel{}
	model.FromDomain(session)

	return r.db.Save(model).Error
}

// Delete はセッションを削除
func (r *SessionRepositoryImpl) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&SessionModel{}).Error
}

// DeleteByUserID はユーザーの全セッションを削除（ログアウト）
func (r *SessionRepositoryImpl) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&SessionModel{}).Error
}

// DeleteExpired は期限切れセッションを削除
func (r *SessionRepositoryImpl) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&SessionModel{}).Error
}
