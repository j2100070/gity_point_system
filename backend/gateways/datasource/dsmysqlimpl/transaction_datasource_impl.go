package dsmysqlimpl

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONB は PostgreSQL の JSONB 型を扱うための型
type JSONB map[string]interface{}

// Value は GORM が DB に書き込む際の変換
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan は DB から読み込む際の変換
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// TransactionModel はGORM用のトランザクションモデル
type TransactionModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FromUserID      *uuid.UUID `gorm:"type:uuid;index"`
	ToUserID        *uuid.UUID `gorm:"type:uuid;index"`
	Amount          int64      `gorm:"not null"`
	TransactionType string     `gorm:"type:varchar(50);not null;index"`
	Status          string     `gorm:"type:varchar(50);not null;index"`
	IdempotencyKey  *string    `gorm:"type:varchar(255);uniqueIndex"`
	Description     string     `gorm:"type:text"`
	Metadata        JSONB      `gorm:"type:jsonb"`
	CreatedAt       time.Time  `gorm:"not null;default:now();index"`
	CompletedAt     *time.Time
}

// TableName はテーブル名を指定
func (TransactionModel) TableName() string {
	return "transactions"
}

// ToDomain はドメインモデルに変換
func (t *TransactionModel) ToDomain() *entities.Transaction {
	return &entities.Transaction{
		ID:              t.ID,
		FromUserID:      t.FromUserID,
		ToUserID:        t.ToUserID,
		Amount:          t.Amount,
		TransactionType: entities.TransactionType(t.TransactionType),
		Status:          entities.TransactionStatus(t.Status),
		IdempotencyKey:  t.IdempotencyKey,
		Description:     t.Description,
		Metadata:        map[string]interface{}(t.Metadata),
		CreatedAt:       t.CreatedAt,
		CompletedAt:     t.CompletedAt,
	}
}

// FromDomain はドメインモデルから変換
func (t *TransactionModel) FromDomain(transaction *entities.Transaction) {
	t.ID = transaction.ID
	t.FromUserID = transaction.FromUserID
	t.ToUserID = transaction.ToUserID
	t.Amount = transaction.Amount
	t.TransactionType = string(transaction.TransactionType)
	t.Status = string(transaction.Status)
	t.IdempotencyKey = transaction.IdempotencyKey
	t.Description = transaction.Description
	t.Metadata = JSONB(transaction.Metadata)
	t.CreatedAt = transaction.CreatedAt
	t.CompletedAt = transaction.CompletedAt
}

// TransactionDataSourceImpl はTransactionDataSourceの実装
type TransactionDataSourceImpl struct {
	db inframysql.DB
}

// NewTransactionDataSource は新しいTransactionDataSourceを作成
func NewTransactionDataSource(db inframysql.DB) dsmysql.TransactionDataSource {
	return &TransactionDataSourceImpl{db: db}
}

// Insert は新しいトランザクションを挿入
func (ds *TransactionDataSourceImpl) Insert(ctx context.Context, transaction *entities.Transaction) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := &TransactionModel{}
	model.FromDomain(transaction)

	if err := db.Create(model).Error; err != nil {
		return err
	}

	*transaction = *model.ToDomain()
	return nil
}

// Select はIDでトランザクションを検索
func (ds *TransactionDataSourceImpl) Select(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	var model TransactionModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByIdempotencyKey は冪等性キーでトランザクションを検索
func (ds *TransactionDataSourceImpl) SelectByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error) {
	var model TransactionModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("idempotency_key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectListByUserID はユーザーに関連するトランザクション一覧を取得
func (ds *TransactionDataSourceImpl) SelectListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error) {
	var models []TransactionModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = model.ToDomain()
	}

	return transactions, nil
}

// SelectListAll は全トランザクション一覧を取得
func (ds *TransactionDataSourceImpl) SelectListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error) {
	var models []TransactionModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = model.ToDomain()
	}

	return transactions, nil
}

// applyFilterConditions はフィルタ条件を適用するヘルパー
func (ds *TransactionDataSourceImpl) applyFilterConditions(query *gorm.DB, transactionType, dateFrom, dateTo string) *gorm.DB {
	if transactionType != "" {
		query = query.Where("transaction_type = ?", transactionType)
	}
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom+" 00:00:00")
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo+" 23:59:59")
	}
	return query
}

// SelectListAllWithFilter はフィルタ・ソート付きで全トランザクション一覧を取得
func (ds *TransactionDataSourceImpl) SelectListAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.Transaction, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	query := db.Model(&TransactionModel{})

	query = ds.applyFilterConditions(query, transactionType, dateFrom, dateTo)

	// ソート（ホワイトリスト方式）
	allowedSortColumns := map[string]string{
		"created_at": "created_at",
		"amount":     "amount",
	}
	col, ok := allowedSortColumns[sortBy]
	if !ok {
		col = "created_at"
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	query = query.Order(col + " " + order)

	var models []TransactionModel
	err := query.Offset(offset).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = model.ToDomain()
	}
	return transactions, nil
}

// CountAll は全トランザクション総数を取得
func (ds *TransactionDataSourceImpl) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := inframysql.GetDB(ctx, ds.db.GetDB()).Model(&TransactionModel{}).Count(&count).Error
	return count, err
}

// CountAllWithFilter はフィルタ付きで全トランザクション総数を取得
func (ds *TransactionDataSourceImpl) CountAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo string) (int64, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	query := db.Model(&TransactionModel{})
	query = ds.applyFilterConditions(query, transactionType, dateFrom, dateTo)
	var count int64
	err := query.Count(&count).Error
	return count, err
}

// Update はトランザクションを更新
func (ds *TransactionDataSourceImpl) Update(ctx context.Context, transaction *entities.Transaction) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := &TransactionModel{}
	model.FromDomain(transaction)

	return db.Model(&TransactionModel{}).
		Where("id = ?", transaction.ID).
		Updates(map[string]interface{}{
			"status":       model.Status,
			"completed_at": model.CompletedAt,
		}).Error
}

// CountByUserID はユーザーのトランザクション総数を取得
func (ds *TransactionDataSourceImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := inframysql.GetDB(ctx, ds.db.GetDB()).Model(&TransactionModel{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Count(&count).Error
	return count, err
}

// IdempotencyKeyModel はGORM用の冪等性キーモデル
type IdempotencyKeyModel struct {
	Key           string     `gorm:"type:varchar(255);primary_key"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	TransactionID *uuid.UUID `gorm:"type:uuid"`
	Status        string     `gorm:"type:varchar(50);not null"`
	CreatedAt     time.Time  `gorm:"not null;default:now()"`
	ExpiresAt     time.Time  `gorm:"not null;index"`
}

// TableName はテーブル名を指定
func (IdempotencyKeyModel) TableName() string {
	return "idempotency_keys"
}

// ToDomain はドメインモデルに変換
func (i *IdempotencyKeyModel) ToDomain() *entities.IdempotencyKey {
	return &entities.IdempotencyKey{
		Key:           i.Key,
		UserID:        i.UserID,
		TransactionID: i.TransactionID,
		Status:        i.Status,
		CreatedAt:     i.CreatedAt,
		ExpiresAt:     i.ExpiresAt,
	}
}

// FromDomain はドメインモデルから変換
func (i *IdempotencyKeyModel) FromDomain(key *entities.IdempotencyKey) {
	i.Key = key.Key
	i.UserID = key.UserID
	i.TransactionID = key.TransactionID
	i.Status = key.Status
	i.CreatedAt = key.CreatedAt
	i.ExpiresAt = key.ExpiresAt
}

// IdempotencyKeyDataSourceImpl はIdempotencyKeyDataSourceの実装
type IdempotencyKeyDataSourceImpl struct {
	db inframysql.DB
}

// NewIdempotencyKeyDataSource は新しいIdempotencyKeyDataSourceを作成
func NewIdempotencyKeyDataSource(db inframysql.DB) dsmysql.IdempotencyKeyDataSource {
	return &IdempotencyKeyDataSourceImpl{db: db}
}

// Insert は新しい冪等性キーを挿入
func (ds *IdempotencyKeyDataSourceImpl) Insert(ctx context.Context, key *entities.IdempotencyKey) error {
	model := &IdempotencyKeyModel{}
	model.FromDomain(key)

	if err := inframysql.GetDB(ctx, ds.db.GetDB()).Create(model).Error; err != nil {
		return err
	}

	*key = *model.ToDomain()
	return nil
}

// SelectByKey はキーで冪等性キーを検索
func (ds *IdempotencyKeyDataSourceImpl) SelectByKey(ctx context.Context, key string) (*entities.IdempotencyKey, error) {
	var model IdempotencyKeyModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("idempotency key not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update は冪等性キーを更新
func (ds *IdempotencyKeyDataSourceImpl) Update(ctx context.Context, key *entities.IdempotencyKey) error {
	model := &IdempotencyKeyModel{}
	model.FromDomain(key)

	return inframysql.GetDB(ctx, ds.db.GetDB()).Model(&IdempotencyKeyModel{}).
		Where("key = ?", key.Key).
		Updates(map[string]interface{}{
			"transaction_id": model.TransactionID,
			"status":         model.Status,
		}).Error
}

// DeleteExpired は期限切れの冪等性キーを削除
func (ds *IdempotencyKeyDataSourceImpl) DeleteExpired(ctx context.Context) error {
	return inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("expires_at < ?", time.Now()).
		Delete(&IdempotencyKeyModel{}).Error
}
