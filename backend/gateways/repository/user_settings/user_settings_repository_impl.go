package user_settings

import (
	"context"

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

// UpdateProfile はプロフィール情報を更新（部分更新、楽観的ロックなし）
func (r *UserSettingsRepositoryImpl) UpdateProfile(ctx context.Context, user *entities.User) (bool, error) {
	r.logger.Debug("Updating user profile", entities.NewField("user_id", user.ID))
	// プロフィール更新では楽観的ロックを使わず、変更されたフィールドのみ更新
	// これにより、画像アップロード後のバージョン競合を回避
	return r.userDS.UpdatePartial(ctx, user.ID, map[string]interface{}{
		"display_name":      user.DisplayName,
		"email":             user.Email,
		"email_verified":    user.EmailVerified,
		"email_verified_at": user.EmailVerifiedAt,
		"avatar_url":        user.AvatarURL,
		"avatar_type":       user.AvatarType,
	})
}

// UpdateUsername はユーザー名を更新（一意性チェック付き）
func (r *UserSettingsRepositoryImpl) UpdateUsername(ctx context.Context, user *entities.User) (bool, error) {
	r.logger.Debug("Updating username", entities.NewField("user_id", user.ID), entities.NewField("new_username", user.Username))
	return r.userDS.Update(ctx, user)
}

// UpdatePassword はパスワードを更新
func (r *UserSettingsRepositoryImpl) UpdatePassword(ctx context.Context, user *entities.User) (bool, error) {
	r.logger.Debug("Updating password", entities.NewField("user_id", user.ID))
	return r.userDS.Update(ctx, user)
}

// CheckUsernameExists はユーザー名が既に存在するかチェック
func (r *UserSettingsRepositoryImpl) CheckUsernameExists(ctx context.Context, username string, excludeUserID uuid.UUID) (bool, error) {
	user, err := r.userDS.SelectByUsername(ctx, username)
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
func (r *UserSettingsRepositoryImpl) CheckEmailExists(ctx context.Context, email string, excludeUserID uuid.UUID) (bool, error) {
	user, err := r.userDS.SelectByEmail(ctx, email)
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
