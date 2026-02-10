package user_settings

import (
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// UsernameChangeHistoryRepositoryImpl はUsernameChangeHistoryRepositoryの実装
type UsernameChangeHistoryRepositoryImpl struct {
	usernameChangeHistoryDS dsmysql.UsernameChangeHistoryDataSource
	logger                  entities.Logger
}

// NewUsernameChangeHistoryRepository は新しいUsernameChangeHistoryRepositoryを作成
func NewUsernameChangeHistoryRepository(usernameChangeHistoryDS dsmysql.UsernameChangeHistoryDataSource, logger entities.Logger) repository.UsernameChangeHistoryRepository {
	return &UsernameChangeHistoryRepositoryImpl{
		usernameChangeHistoryDS: usernameChangeHistoryDS,
		logger:                  logger,
	}
}

// Create は新しいユーザー名変更履歴を作成
func (r *UsernameChangeHistoryRepositoryImpl) Create(history *entities.UsernameChangeHistory) error {
	r.logger.Debug("Creating username change history",
		entities.NewField("user_id", history.UserID),
		entities.NewField("old_username", history.OldUsername),
		entities.NewField("new_username", history.NewUsername))
	return r.usernameChangeHistoryDS.Insert(history)
}

// ReadListByUserID はユーザーIDで履歴を取得
func (r *UsernameChangeHistoryRepositoryImpl) ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error) {
	return r.usernameChangeHistoryDS.SelectListByUserID(userID, offset, limit)
}

// CountByUserID はユーザーIDで履歴数を取得
func (r *UsernameChangeHistoryRepositoryImpl) CountByUserID(userID uuid.UUID) (int64, error) {
	return r.usernameChangeHistoryDS.CountByUserID(userID)
}
