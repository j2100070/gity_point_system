package transaction

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// RepositoryImpl はTransactionRepositoryの実装
type RepositoryImpl struct {
	transactionDS dsmysql.TransactionDataSource
	logger        entities.Logger
}

// NewTransactionRepository は新しいTransactionRepositoryを作成
func NewTransactionRepository(
	transactionDS dsmysql.TransactionDataSource,
	logger entities.Logger,
) repository.TransactionRepository {
	return &RepositoryImpl{
		transactionDS: transactionDS,
		logger:        logger,
	}
}

// Create は新しいトランザクションを作成
func (r *RepositoryImpl) Create(ctx context.Context, transaction *entities.Transaction) error {
	r.logger.Debug("Creating transaction", entities.NewField("transaction_id", transaction.ID))
	return r.transactionDS.Insert(ctx, transaction)
}

// Read はIDでトランザクションを検索
func (r *RepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	return r.transactionDS.Select(ctx, id)
}

// ReadByIdempotencyKey は冪等性キーでトランザクションを検索
func (r *RepositoryImpl) ReadByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error) {
	return r.transactionDS.SelectByIdempotencyKey(ctx, key)
}

// ReadListByUserID はユーザーに関連するトランザクション一覧を取得
func (r *RepositoryImpl) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error) {
	return r.transactionDS.SelectListByUserID(ctx, userID, offset, limit)
}

// ReadListAll は全トランザクション一覧を取得
func (r *RepositoryImpl) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error) {
	return r.transactionDS.SelectListAll(ctx, offset, limit)
}

// ReadListAllWithFilter はフィルタ・ソート付きで全トランザクション一覧を取得
func (r *RepositoryImpl) ReadListAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.Transaction, error) {
	return r.transactionDS.SelectListAllWithFilter(ctx, transactionType, dateFrom, dateTo, sortBy, sortOrder, offset, limit)
}

// CountAll は全トランザクション総数を取得
func (r *RepositoryImpl) CountAll(ctx context.Context) (int64, error) {
	return r.transactionDS.CountAll(ctx)
}

// CountAllWithFilter はフィルタ付きで全トランザクション総数を取得
func (r *RepositoryImpl) CountAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo string) (int64, error) {
	return r.transactionDS.CountAllWithFilter(ctx, transactionType, dateFrom, dateTo)
}

// Update はトランザクションを更新
func (r *RepositoryImpl) Update(ctx context.Context, transaction *entities.Transaction) error {
	r.logger.Debug("Updating transaction", entities.NewField("transaction_id", transaction.ID))
	return r.transactionDS.Update(ctx, transaction)
}

// CountByUserID はユーザーのトランザクション総数を取得
func (r *RepositoryImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.transactionDS.CountByUserID(ctx, userID)
}

// IdempotencyRepositoryImpl はIdempotencyKeyRepositoryの実装
type IdempotencyRepositoryImpl struct {
	idempotencyDS dsmysql.IdempotencyKeyDataSource
	logger        entities.Logger
}

// NewIdempotencyKeyRepository は新しいIdempotencyKeyRepositoryを作成
func NewIdempotencyKeyRepository(
	idempotencyDS dsmysql.IdempotencyKeyDataSource,
	logger entities.Logger,
) repository.IdempotencyKeyRepository {
	return &IdempotencyRepositoryImpl{
		idempotencyDS: idempotencyDS,
		logger:        logger,
	}
}

// Create は新しい冪等性キーを作成
func (r *IdempotencyRepositoryImpl) Create(ctx context.Context, key *entities.IdempotencyKey) error {
	r.logger.Debug("Creating idempotency key", entities.NewField("key", key.Key))
	return r.idempotencyDS.Insert(ctx, key)
}

// ReadByKey はキーで冪等性キーを検索
func (r *IdempotencyRepositoryImpl) ReadByKey(ctx context.Context, key string) (*entities.IdempotencyKey, error) {
	return r.idempotencyDS.SelectByKey(ctx, key)
}

// Update は冪等性キーを更新
func (r *IdempotencyRepositoryImpl) Update(ctx context.Context, key *entities.IdempotencyKey) error {
	r.logger.Debug("Updating idempotency key", entities.NewField("key", key.Key))
	return r.idempotencyDS.Update(ctx, key)
}

// DeleteExpired は期限切れの冪等性キーを削除
func (r *IdempotencyRepositoryImpl) DeleteExpired(ctx context.Context) error {
	r.logger.Debug("Deleting expired idempotency keys")
	return r.idempotencyDS.DeleteExpired(ctx)
}
