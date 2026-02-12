package user_settings

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// ArchivedUserRepositoryImpl はArchivedUserRepositoryの実装
type ArchivedUserRepositoryImpl struct {
	archivedUserDS dsmysql.ArchivedUserDataSource
	logger         entities.Logger
}

// NewArchivedUserRepository は新しいArchivedUserRepositoryを作成
func NewArchivedUserRepository(archivedUserDS dsmysql.ArchivedUserDataSource, logger entities.Logger) repository.ArchivedUserRepository {
	return &ArchivedUserRepositoryImpl{
		archivedUserDS: archivedUserDS,
		logger:         logger,
	}
}

// Create はアーカイブユーザーを作成
func (r *ArchivedUserRepositoryImpl) Create(ctx context.Context, archivedUser *entities.ArchivedUser) error {
	r.logger.Debug("Creating archived user", entities.NewField("user_id", archivedUser.ID))
	return r.archivedUserDS.Insert(ctx, archivedUser)
}

// Read はIDでアーカイブユーザーを検索
func (r *ArchivedUserRepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.ArchivedUser, error) {
	return r.archivedUserDS.Select(ctx, id)
}

// ReadByUsername はユーザー名でアーカイブユーザーを検索
func (r *ArchivedUserRepositoryImpl) ReadByUsername(ctx context.Context, username string) (*entities.ArchivedUser, error) {
	return r.archivedUserDS.SelectByUsername(ctx, username)
}

// ReadList はアーカイブユーザー一覧を取得
func (r *ArchivedUserRepositoryImpl) ReadList(ctx context.Context, offset, limit int) ([]*entities.ArchivedUser, error) {
	return r.archivedUserDS.SelectList(ctx, offset, limit)
}

// Count はアーカイブユーザー総数を取得
func (r *ArchivedUserRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.archivedUserDS.Count(ctx)
}

// Restore はアーカイブユーザーを復元（アーカイブから削除してユーザーに戻す）
func (r *ArchivedUserRepositoryImpl) Restore(ctx context.Context, tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error {
	r.logger.Debug("Restoring archived user", entities.NewField("user_id", archivedUser.ID))
	return r.archivedUserDS.Restore(ctx, tx, archivedUser, user)
}
