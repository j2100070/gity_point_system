package persistence

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionModel はGORM用のトランザクションモデル
type TransactionModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FromUserID      *uuid.UUID `gorm:"type:uuid;index"`
	ToUserID        *uuid.UUID `gorm:"type:uuid;index"`
	Amount          int64      `gorm:"not null;check:amount > 0"`
	TransactionType string     `gorm:"type:varchar(50);not null;index"`
	Status          string     `gorm:"type:varchar(50);not null;default:'completed'"`
	IdempotencyKey  *string    `gorm:"type:varchar(255);index"`
	Description     string     `gorm:"type:text"`
	Metadata        string     `gorm:"type:jsonb"` // JSONBとして保存
	CreatedAt       time.Time  `gorm:"not null;default:now();index:,sort:desc"`
	CompletedAt     *time.Time
}

// TableName はテーブル名を指定
func (TransactionModel) TableName() string {
	return "transactions"
}

// ToDomain はドメインモデルに変換
func (t *TransactionModel) ToDomain() (*domain.Transaction, error) {
	var metadata map[string]interface{}
	if t.Metadata != "" {
		if err := json.Unmarshal([]byte(t.Metadata), &metadata); err != nil {
			return nil, err
		}
	} else {
		metadata = make(map[string]interface{})
	}

	return &domain.Transaction{
		ID:              t.ID,
		FromUserID:      t.FromUserID,
		ToUserID:        t.ToUserID,
		Amount:          t.Amount,
		TransactionType: domain.TransactionType(t.TransactionType),
		Status:          domain.TransactionStatus(t.Status),
		IdempotencyKey:  t.IdempotencyKey,
		Description:     t.Description,
		Metadata:        metadata,
		CreatedAt:       t.CreatedAt,
		CompletedAt:     t.CompletedAt,
	}, nil
}

// FromDomain はドメインモデルから変換
func (t *TransactionModel) FromDomain(transaction *domain.Transaction) error {
	metadataBytes, err := json.Marshal(transaction.Metadata)
	if err != nil {
		return err
	}

	t.ID = transaction.ID
	t.FromUserID = transaction.FromUserID
	t.ToUserID = transaction.ToUserID
	t.Amount = transaction.Amount
	t.TransactionType = string(transaction.TransactionType)
	t.Status = string(transaction.Status)
	t.IdempotencyKey = transaction.IdempotencyKey
	t.Description = transaction.Description
	t.Metadata = string(metadataBytes)
	t.CreatedAt = transaction.CreatedAt
	t.CompletedAt = transaction.CompletedAt

	return nil
}

// TransactionRepositoryImpl はTransactionRepositoryの実装
type TransactionRepositoryImpl struct {
	db *gorm.DB
}

// NewTransactionRepository は新しいTransactionRepositoryを作成
func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &TransactionRepositoryImpl{db: db}
}

// Create は新しいトランザクションを作成
func (r *TransactionRepositoryImpl) Create(tx interface{}, transaction *domain.Transaction) error {
	db := r.db
	if tx != nil {
		db = tx.(*gorm.DB)
	}

	model := &TransactionModel{}
	if err := model.FromDomain(transaction); err != nil {
		return err
	}

	if err := db.Create(model).Error; err != nil {
		return err
	}

	result, err := model.ToDomain()
	if err != nil {
		return err
	}
	*transaction = *result

	return nil
}

// FindByID はIDでトランザクションを検索
func (r *TransactionRepositoryImpl) FindByID(id uuid.UUID) (*domain.Transaction, error) {
	var model TransactionModel

	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByIdempotencyKey は冪等性キーでトランザクションを検索
func (r *TransactionRepositoryImpl) FindByIdempotencyKey(key string) (*domain.Transaction, error) {
	var model TransactionModel

	err := r.db.Where("idempotency_key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// ListByUserID はユーザーに関連するトランザクション一覧を取得
func (r *TransactionRepositoryImpl) ListByUserID(userID uuid.UUID, offset, limit int) ([]*domain.Transaction, error) {
	var models []TransactionModel

	err := r.db.Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.Transaction, len(models))
	for i, model := range models {
		t, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		transactions[i] = t
	}

	return transactions, nil
}

// ListAll は全トランザクション一覧を取得（管理者用）
func (r *TransactionRepositoryImpl) ListAll(offset, limit int) ([]*domain.Transaction, error) {
	var models []TransactionModel

	err := r.db.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.Transaction, len(models))
	for i, model := range models {
		t, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		transactions[i] = t
	}

	return transactions, nil
}

// Update はトランザクションを更新
func (r *TransactionRepositoryImpl) Update(tx interface{}, transaction *domain.Transaction) error {
	db := r.db
	if tx != nil {
		db = tx.(*gorm.DB)
	}

	model := &TransactionModel{}
	if err := model.FromDomain(transaction); err != nil {
		return err
	}

	return db.Save(model).Error
}

// CountByUserID はユーザーのトランザクション総数を取得
func (r *TransactionRepositoryImpl) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&TransactionModel{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Count(&count).Error
	return count, err
}

// IdempotencyKeyModel はGORM用の冪等性キーモデル
type IdempotencyKeyModel struct {
	Key           string     `gorm:"type:varchar(255);primary_key"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	TransactionID *uuid.UUID `gorm:"type:uuid"`
	Status        string     `gorm:"type:varchar(50);not null;default:'processing'"`
	CreatedAt     time.Time  `gorm:"not null;default:now()"`
	ExpiresAt     time.Time  `gorm:"not null;index"`
}

// TableName はテーブル名を指定
func (IdempotencyKeyModel) TableName() string {
	return "idempotency_keys"
}

// ToDomain はドメインモデルに変換
func (i *IdempotencyKeyModel) ToDomain() *domain.IdempotencyKey {
	return &domain.IdempotencyKey{
		Key:           i.Key,
		UserID:        i.UserID,
		TransactionID: i.TransactionID,
		Status:        i.Status,
		CreatedAt:     i.CreatedAt,
		ExpiresAt:     i.ExpiresAt,
	}
}

// FromDomain はドメインモデルから変換
func (i *IdempotencyKeyModel) FromDomain(key *domain.IdempotencyKey) {
	i.Key = key.Key
	i.UserID = key.UserID
	i.TransactionID = key.TransactionID
	i.Status = key.Status
	i.CreatedAt = key.CreatedAt
	i.ExpiresAt = key.ExpiresAt
}

// IdempotencyKeyRepositoryImpl はIdempotencyKeyRepositoryの実装
type IdempotencyKeyRepositoryImpl struct {
	db *gorm.DB
}

// NewIdempotencyKeyRepository は新しいIdempotencyKeyRepositoryを作成
func NewIdempotencyKeyRepository(db *gorm.DB) domain.IdempotencyKeyRepository {
	return &IdempotencyKeyRepositoryImpl{db: db}
}

// Create は新しい冪等性キーを作成（既存の場合はエラー）
func (r *IdempotencyKeyRepositoryImpl) Create(key *domain.IdempotencyKey) error {
	model := &IdempotencyKeyModel{}
	model.FromDomain(key)

	// 既存チェック（競合エラーを利用）
	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	*key = *model.ToDomain()
	return nil
}

// FindByKey はキーで冪等性キーを検索
func (r *IdempotencyKeyRepositoryImpl) FindByKey(key string) (*domain.IdempotencyKey, error) {
	var model IdempotencyKeyModel

	err := r.db.Where("key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("idempotency key not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update は冪等性キーを更新
func (r *IdempotencyKeyRepositoryImpl) Update(key *domain.IdempotencyKey) error {
	model := &IdempotencyKeyModel{}
	model.FromDomain(key)

	return r.db.Save(model).Error
}

// DeleteExpired は期限切れの冪等性キーを削除
func (r *IdempotencyKeyRepositoryImpl) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&IdempotencyKeyModel{}).Error
}
