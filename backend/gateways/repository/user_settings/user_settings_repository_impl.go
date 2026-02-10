package user_settings

import (
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// UserSettingsRepositoryImpl はUserSettingsRepositoryの実装
type UserSettingsRepositoryImpl struct {
	userDS dsmysql.UserDataSource
	logger entities.Logger
}

// NewUserSettingsRepository は新しいUserSettingsRepositoryを作成
func NewUserSettingsRepository(userDS dsmysql.UserDataSource, logger entities.Logger) repository.UserSettingsRepository {
	return &UserSettingsRepositoryImpl{
		userDS: userDS,
		logger: logger,
	}
}

// UpdateProfile はプロフィール情報を更新（楽観的ロック対応）
func (r *UserSettingsRepositoryImpl) UpdateProfile(user *entities.User) (bool, error) {
	r.logger.Debug("Updating user profile", entities.NewField("user_id", user.ID))
	return r.userDS.Update(user)
}

// UpdateUsername はユーザー名を更新（一意性チェック付き）
func (r *UserSettingsRepositoryImpl) UpdateUsername(user *entities.User) (bool, error) {
	r.logger.Debug("Updating username", entities.NewField("user_id", user.ID), entities.NewField("new_username", user.Username))
	return r.userDS.Update(user)
}

// UpdatePassword はパスワードを更新
func (r *UserSettingsRepositoryImpl) UpdatePassword(user *entities.User) (bool, error) {
	r.logger.Debug("Updating password", entities.NewField("user_id", user.ID))
	return r.userDS.Update(user)
}

// CheckUsernameExists はユーザー名が既に存在するかチェック
func (r *UserSettingsRepositoryImpl) CheckUsernameExists(username string, excludeUserID uuid.UUID) (bool, error) {
	user, err := r.userDS.SelectByUsername(username)
	if err != nil {
		// ユーザーが見つからない場合は存在しない
		return false, nil
	}
	// 除外するユーザーIDと一致する場合は存在しないとみなす
	if user.ID == excludeUserID {
		return false, nil
	}
	return true, nil
}

// CheckEmailExists はメールアドレスが既に存在するかチェック
func (r *UserSettingsRepositoryImpl) CheckEmailExists(email string, excludeUserID uuid.UUID) (bool, error) {
	user, err := r.userDS.SelectByEmail(email)
	if err != nil {
		// ユーザーが見つからない場合は存在しない
		return false, nil
	}
	// 除外するユーザーIDと一致する場合は存在しないとみなす
	if user.ID == excludeUserID {
		return false, nil
	}
	return true, nil
}
