package transaction

import (
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
func (r *RepositoryImpl) Create(tx interface{}, transaction *entities.Transaction) error {
	r.logger.Debug("Creating transaction", entities.NewField("transaction_id", transaction.ID))
	return r.transactionDS.Insert(tx, transaction)
}

// Read はIDでトランザクションを検索
func (r *RepositoryImpl) Read(id uuid.UUID) (*entities.Transaction, error) {
	return r.transactionDS.Select(id)
}

// ReadByIdempotencyKey は冪等性キーでトランザクションを検索
func (r *RepositoryImpl) ReadByIdempotencyKey(key string) (*entities.Transaction, error) {
	return r.transactionDS.SelectByIdempotencyKey(key)
}

// ReadListByUserID はユーザーに関連するトランザクション一覧を取得
func (r *RepositoryImpl) ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error) {
	return r.transactionDS.SelectListByUserID(userID, offset, limit)
}

// ReadListAll は全トランザクション一覧を取得
func (r *RepositoryImpl) ReadListAll(offset, limit int) ([]*entities.Transaction, error) {
	return r.transactionDS.SelectListAll(offset, limit)
}

// Update はトランザクションを更新
func (r *RepositoryImpl) Update(tx interface{}, transaction *entities.Transaction) error {
	r.logger.Debug("Updating transaction", entities.NewField("transaction_id", transaction.ID))
	return r.transactionDS.Update(tx, transaction)
}

// CountByUserID はユーザーのトランザクション総数を取得
func (r *RepositoryImpl) CountByUserID(userID uuid.UUID) (int64, error) {
	return r.transactionDS.CountByUserID(userID)
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
func (r *IdempotencyRepositoryImpl) Create(key *entities.IdempotencyKey) error {
	r.logger.Debug("Creating idempotency key", entities.NewField("key", key.Key))
	return r.idempotencyDS.Insert(key)
}

// ReadByKey はキーで冪等性キーを検索
func (r *IdempotencyRepositoryImpl) ReadByKey(key string) (*entities.IdempotencyKey, error) {
	return r.idempotencyDS.SelectByKey(key)
}

// Update は冪等性キーを更新
func (r *IdempotencyRepositoryImpl) Update(key *entities.IdempotencyKey) error {
	r.logger.Debug("Updating idempotency key", entities.NewField("key", key.Key))
	return r.idempotencyDS.Update(key)
}

// DeleteExpired は期限切れの冪等性キーを削除
func (r *IdempotencyRepositoryImpl) DeleteExpired() error {
	r.logger.Debug("Deleting expired idempotency keys")
	return r.idempotencyDS.DeleteExpired()
}
