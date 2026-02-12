package user_settings

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// PasswordChangeHistoryRepositoryImpl はPasswordChangeHistoryRepositoryの実装
type PasswordChangeHistoryRepositoryImpl struct {
	passwordChangeHistoryDS dsmysql.PasswordChangeHistoryDataSource
	logger                  entities.Logger
}

// NewPasswordChangeHistoryRepository は新しいPasswordChangeHistoryRepositoryを作成
func NewPasswordChangeHistoryRepository(passwordChangeHistoryDS dsmysql.PasswordChangeHistoryDataSource, logger entities.Logger) repository.PasswordChangeHistoryRepository {
	return &PasswordChangeHistoryRepositoryImpl{
		passwordChangeHistoryDS: passwordChangeHistoryDS,
		logger:                  logger,
	}
}

// Create は新しいパスワード変更履歴を作成
func (r *PasswordChangeHistoryRepositoryImpl) Create(ctx context.Context, history *entities.PasswordChangeHistory) error {
	r.logger.Debug("Creating password change history", entities.NewField("user_id", history.UserID))
	return r.passwordChangeHistoryDS.Insert(ctx, history)
}

// ReadListByUserID はユーザーIDで履歴を取得
func (r *PasswordChangeHistoryRepositoryImpl) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error) {
	return r.passwordChangeHistoryDS.SelectListByUserID(ctx, userID, offset, limit)
}

// CountByUserID はユーザーIDで履歴数を取得
func (r *PasswordChangeHistoryRepositoryImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.passwordChangeHistoryDS.CountByUserID(ctx, userID)
}
