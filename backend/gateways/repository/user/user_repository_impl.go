package user

import (
	"context"

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
func (r *RepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	r.logger.Debug("Creating user", entities.NewField("username", user.Username))
	return r.userDS.Insert(ctx, user)
}

// Read はIDでユーザーを検索
func (r *RepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	// contextが渡された場合はcontextを使用、nilの場合はレガシーメソッド
	if ctx != nil {
		return r.userDS.Select(ctx, id)
	}
	// Fallback for legacy code
	return r.userDS.Select(ctx, id)
}

// ReadByUsername はユーザー名でユーザーを検索
func (r *RepositoryImpl) ReadByUsername(ctx context.Context, username string) (*entities.User, error) {
	return r.userDS.SelectByUsername(ctx, username)
}

// ReadByEmail はメールアドレスでユーザーを検索
func (r *RepositoryImpl) ReadByEmail(ctx context.Context, email string) (*entities.User, error) {
	return r.userDS.SelectByEmail(ctx, email)
}

// Update はユーザー情報を更新（楽観的ロック対応）
func (r *RepositoryImpl) Update(ctx context.Context, user *entities.User) (bool, error) {
	r.logger.Debug("Updating user", entities.NewField("user_id", user.ID))
	return r.userDS.Update(ctx, user)
}

// UpdateBalanceWithLock は残高を更新（悲観的ロック）
func (r *RepositoryImpl) UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error {
	r.logger.Debug("Updating balance with lock",
		entities.NewField("user_id", userID),
		entities.NewField("amount", amount),
		entities.NewField("is_deduct", isDeduct))
	return r.userDS.UpdateBalanceWithLock(ctx, userID, amount, isDeduct)
}

// UpdateBalancesWithLock は複数ユーザーの残高を一括更新（悲観的ロック、デッドロック回避）
func (r *RepositoryImpl) UpdateBalancesWithLock(ctx context.Context, updates []repository.BalanceUpdate) error {
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

	return r.userDS.UpdateBalancesWithLock(ctx, dsUpdates)
}

// ReadList はユーザー一覧を取得
func (r *RepositoryImpl) ReadList(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	return r.userDS.SelectList(ctx, offset, limit)
}

// ReadListWithSearch は検索・ソート付きでユーザー一覧を取得
func (r *RepositoryImpl) ReadListWithSearch(ctx context.Context, search, sortBy, sortOrder string, offset, limit int) ([]*entities.User, error) {
	return r.userDS.SelectListWithSearch(ctx, search, sortBy, sortOrder, offset, limit)
}

// Count はユーザー総数を取得
func (r *RepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.userDS.Count(ctx)
}

// CountWithSearch は検索条件付きでユーザー総数を取得
func (r *RepositoryImpl) CountWithSearch(ctx context.Context, search string) (int64, error) {
	return r.userDS.CountWithSearch(ctx, search)
}

// Delete はユーザーを論理削除
func (r *RepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting user", entities.NewField("user_id", id))
	return r.userDS.Delete(ctx, id)
}
