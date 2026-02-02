package session

import (
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// RepositoryImpl はSessionRepositoryの実装
type RepositoryImpl struct {
	sessionDS dsmysql.SessionDataSource
	logger    entities.Logger
}

// NewSessionRepository は新しいSessionRepositoryを作成
func NewSessionRepository(
	sessionDS dsmysql.SessionDataSource,
	logger entities.Logger,
) repository.SessionRepository {
	return &RepositoryImpl{
		sessionDS: sessionDS,
		logger:    logger,
	}
}

// Create は新しいセッションを作成
func (r *RepositoryImpl) Create(session *entities.Session) error {
	r.logger.Debug("Creating session", entities.NewField("user_id", session.UserID))
	return r.sessionDS.Insert(session)
}

// ReadByToken はトークンでセッションを検索
func (r *RepositoryImpl) ReadByToken(token string) (*entities.Session, error) {
	return r.sessionDS.SelectByToken(token)
}

// Update はセッションを更新
func (r *RepositoryImpl) Update(session *entities.Session) error {
	r.logger.Debug("Updating session", entities.NewField("session_id", session.ID))
	return r.sessionDS.Update(session)
}

// Delete はセッションを削除
func (r *RepositoryImpl) Delete(id uuid.UUID) error {
	r.logger.Debug("Deleting session", entities.NewField("session_id", id))
	return r.sessionDS.Delete(id)
}

// DeleteByUserID はユーザーの全セッションを削除
func (r *RepositoryImpl) DeleteByUserID(userID uuid.UUID) error {
	r.logger.Debug("Deleting user sessions", entities.NewField("user_id", userID))
	return r.sessionDS.DeleteByUserID(userID)
}

// DeleteExpired は期限切れセッションを削除
func (r *RepositoryImpl) DeleteExpired() error {
	r.logger.Debug("Deleting expired sessions")
	return r.sessionDS.DeleteExpired()
}
