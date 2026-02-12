package user_settings

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// EmailVerificationRepositoryImpl はEmailVerificationRepositoryの実装
type EmailVerificationRepositoryImpl struct {
	emailVerificationDS dsmysql.EmailVerificationDataSource
	logger              entities.Logger
}

// NewEmailVerificationRepository は新しいEmailVerificationRepositoryを作成
func NewEmailVerificationRepository(emailVerificationDS dsmysql.EmailVerificationDataSource, logger entities.Logger) repository.EmailVerificationRepository {
	return &EmailVerificationRepositoryImpl{
		emailVerificationDS: emailVerificationDS,
		logger:              logger,
	}
}

// Create は新しいメール認証トークンを作成
func (r *EmailVerificationRepositoryImpl) Create(ctx context.Context, token *entities.EmailVerificationToken) error {
	r.logger.Debug("Creating email verification token", entities.NewField("email", token.Email), entities.NewField("token_type", token.TokenType))
	return r.emailVerificationDS.Insert(ctx, token)
}

// ReadByToken はトークンで検索
func (r *EmailVerificationRepositoryImpl) ReadByToken(ctx context.Context, token string) (*entities.EmailVerificationToken, error) {
	return r.emailVerificationDS.SelectByToken(ctx, token)
}

// Update はトークン情報を更新
func (r *EmailVerificationRepositoryImpl) Update(ctx context.Context, token *entities.EmailVerificationToken) error {
	r.logger.Debug("Updating email verification token", entities.NewField("token_id", token.ID))
	return r.emailVerificationDS.Update(ctx, token)
}

// DeleteExpired は期限切れのトークンを削除
func (r *EmailVerificationRepositoryImpl) DeleteExpired(ctx context.Context) error {
	r.logger.Debug("Deleting expired email verification tokens")
	return r.emailVerificationDS.DeleteExpired(ctx)
}

// DeleteByUserID はユーザーIDに紐づくトークンを削除
func (r *EmailVerificationRepositoryImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("Deleting email verification tokens by user ID", entities.NewField("user_id", userID))
	return r.emailVerificationDS.DeleteByUserID(ctx, userID)
}
