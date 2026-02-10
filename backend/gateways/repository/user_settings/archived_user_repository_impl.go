package user_settings

import (
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
func (r *ArchivedUserRepositoryImpl) Create(archivedUser *entities.ArchivedUser) error {
	r.logger.Debug("Creating archived user", entities.NewField("user_id", archivedUser.ID))
	return r.archivedUserDS.Insert(archivedUser)
}

// Read はIDでアーカイブユーザーを検索
func (r *ArchivedUserRepositoryImpl) Read(id uuid.UUID) (*entities.ArchivedUser, error) {
	return r.archivedUserDS.Select(id)
}

// ReadByUsername はユーザー名でアーカイブユーザーを検索
func (r *ArchivedUserRepositoryImpl) ReadByUsername(username string) (*entities.ArchivedUser, error) {
	return r.archivedUserDS.SelectByUsername(username)
}

// ReadList はアーカイブユーザー一覧を取得
func (r *ArchivedUserRepositoryImpl) ReadList(offset, limit int) ([]*entities.ArchivedUser, error) {
	return r.archivedUserDS.SelectList(offset, limit)
}

// Count はアーカイブユーザー総数を取得
func (r *ArchivedUserRepositoryImpl) Count() (int64, error) {
	return r.archivedUserDS.Count()
}

// Restore はアーカイブユーザーを復元（アーカイブから削除してユーザーに戻す）
func (r *ArchivedUserRepositoryImpl) Restore(tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error {
	r.logger.Debug("Restoring archived user", entities.NewField("user_id", archivedUser.ID))
	return r.archivedUserDS.Restore(tx, archivedUser, user)
}
