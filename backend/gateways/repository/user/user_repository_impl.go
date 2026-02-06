package user

import (
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// RepositoryImpl はUserRepositoryの実装
// DataSourceを活用し、UseCasesレイヤーが実際のテーブル構造を把握しなくてもEntityの永続化を行える
type RepositoryImpl struct {
	userDS dsmysql.UserDataSource
	logger entities.Logger
}

// NewUserRepository は新しいUserRepositoryを作成
func NewUserRepository(userDS dsmysql.UserDataSource, logger entities.Logger) repository.UserRepository {
	return &RepositoryImpl{
		userDS: userDS,
		logger: logger,
	}
}

// Create は新しいユーザーを作成
func (r *RepositoryImpl) Create(user *entities.User) error {
	r.logger.Debug("Creating user", entities.NewField("username", user.Username))
	return r.userDS.Insert(user)
}

// Read はIDでユーザーを検索
func (r *RepositoryImpl) Read(id uuid.UUID) (*entities.User, error) {
	return r.userDS.Select(id)
}

// ReadByUsername はユーザー名でユーザーを検索
func (r *RepositoryImpl) ReadByUsername(username string) (*entities.User, error) {
	return r.userDS.SelectByUsername(username)
}

// ReadByEmail はメールアドレスでユーザーを検索
func (r *RepositoryImpl) ReadByEmail(email string) (*entities.User, error) {
	return r.userDS.SelectByEmail(email)
}

// Update はユーザー情報を更新（楽観的ロック対応）
func (r *RepositoryImpl) Update(user *entities.User) (bool, error) {
	r.logger.Debug("Updating user", entities.NewField("user_id", user.ID))
	return r.userDS.Update(user)
}

// UpdateBalanceWithLock は残高を更新（悲観的ロック）
func (r *RepositoryImpl) UpdateBalanceWithLock(tx interface{}, userID uuid.UUID, amount int64, isDeduct bool) error {
	r.logger.Debug("Updating balance with lock",
		entities.NewField("user_id", userID),
		entities.NewField("amount", amount),
		entities.NewField("is_deduct", isDeduct))
	return r.userDS.UpdateBalanceWithLock(tx, userID, amount, isDeduct)
}

// UpdateBalancesWithLock は複数ユーザーの残高を一括更新（悲観的ロック、デッドロック回避）
func (r *RepositoryImpl) UpdateBalancesWithLock(tx interface{}, updates []repository.BalanceUpdate) error {
	r.logger.Debug("Updating multiple balances with lock",
		entities.NewField("count", len(updates)))

	// repository.BalanceUpdate を dsmysql.BalanceUpdate に変換
	dsUpdates := make([]dsmysql.BalanceUpdate, len(updates))
	for i, update := range updates {
		dsUpdates[i] = dsmysql.BalanceUpdate{
			UserID:   update.UserID,
			Amount:   update.Amount,
			IsDeduct: update.IsDeduct,
		}
	}

	return r.userDS.UpdateBalancesWithLock(tx, dsUpdates)
}

// ReadList はユーザー一覧を取得
func (r *RepositoryImpl) ReadList(offset, limit int) ([]*entities.User, error) {
	return r.userDS.SelectList(offset, limit)
}

// Count はユーザー総数を取得
func (r *RepositoryImpl) Count() (int64, error) {
	return r.userDS.Count()
}

// Delete はユーザーを論理削除
func (r *RepositoryImpl) Delete(id uuid.UUID) error {
	r.logger.Debug("Deleting user", entities.NewField("user_id", id))
	return r.userDS.Delete(id)
}
