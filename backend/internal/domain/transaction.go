package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TransactionType は取引タイプ
type TransactionType string

const (
	TransactionTypeTransfer     TransactionType = "transfer"      // ユーザー間送金
	TransactionTypeAdminGrant   TransactionType = "admin_grant"   // 管理者付与
	TransactionTypeAdminDeduct  TransactionType = "admin_deduct"  // 管理者減算
	TransactionTypeSystemGrant  TransactionType = "system_grant"  // システム付与
)

// TransactionStatus は取引状態
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

// Transaction はポイント取引エンティティ
type Transaction struct {
	ID              uuid.UUID
	FromUserID      *uuid.UUID        // 送信者（nilの場合はシステム付与）
	ToUserID        *uuid.UUID        // 受信者（nilの場合はシステムへの返却）
	Amount          int64
	TransactionType TransactionType
	Status          TransactionStatus
	IdempotencyKey  *string           // 冪等性キー
	Description     string
	Metadata        map[string]interface{} // 追加メタデータ（JSONBとして保存）
	CreatedAt       time.Time
	CompletedAt     *time.Time
}

// NewTransfer はユーザー間送金トランザクションを作成
func NewTransfer(fromUserID, toUserID uuid.UUID, amount int64, idempotencyKey string, description string) (*Transaction, error) {
	if fromUserID == toUserID {
		return nil, errors.New("cannot transfer to the same user")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	toUserIDPtr := toUserID
	return &Transaction{
		ID:              uuid.New(),
		FromUserID:      &fromUserID,
		ToUserID:        &toUserIDPtr,
		Amount:          amount,
		TransactionType: TransactionTypeTransfer,
		Status:          TransactionStatusPending,
		IdempotencyKey:  &idempotencyKey,
		Description:     description,
		Metadata:        make(map[string]interface{}),
		CreatedAt:       time.Now(),
	}, nil
}

// NewAdminGrant は管理者によるポイント付与トランザクションを作成
func NewAdminGrant(toUserID uuid.UUID, amount int64, description string, adminID uuid.UUID) (*Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	metadata := map[string]interface{}{
		"admin_id": adminID.String(),
	}

	toUserIDPtr := toUserID
	return &Transaction{
		ID:              uuid.New(),
		FromUserID:      nil, // システムからの付与
		ToUserID:        &toUserIDPtr,
		Amount:          amount,
		TransactionType: TransactionTypeAdminGrant,
		Status:          TransactionStatusCompleted,
		Description:     description,
		Metadata:        metadata,
		CreatedAt:       time.Now(),
		CompletedAt:     ptrTime(time.Now()),
	}, nil
}

// NewAdminDeduct は管理者によるポイント減算トランザクションを作成
func NewAdminDeduct(fromUserID uuid.UUID, amount int64, description string, adminID uuid.UUID) (*Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	metadata := map[string]interface{}{
		"admin_id": adminID.String(),
	}

	// 減算は逆方向の取引として記録
	return &Transaction{
		ID:              uuid.New(),
		FromUserID:      &fromUserID,
		ToUserID:        nil, // システムへの返却
		Amount:          amount,
		TransactionType: TransactionTypeAdminDeduct,
		Status:          TransactionStatusCompleted,
		Description:     description,
		Metadata:        metadata,
		CreatedAt:       time.Now(),
		CompletedAt:     ptrTime(time.Now()),
	}, nil
}

// Complete は取引を完了状態にする
func (t *Transaction) Complete() error {
	if t.Status != TransactionStatusPending {
		return errors.New("transaction is not in pending status")
	}
	t.Status = TransactionStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
	return nil
}

// Fail は取引を失敗状態にする
func (t *Transaction) Fail() error {
	if t.Status != TransactionStatusPending {
		return errors.New("transaction is not in pending status")
	}
	t.Status = TransactionStatusFailed
	return nil
}

// TransactionRepository はトランザクションのリポジトリインターフェース
type TransactionRepository interface {
	// Create は新しいトランザクションを作成
	Create(tx interface{}, transaction *Transaction) error

	// FindByID はIDでトランザクションを検索
	FindByID(id uuid.UUID) (*Transaction, error)

	// FindByIdempotencyKey は冪等性キーでトランザクションを検索
	FindByIdempotencyKey(key string) (*Transaction, error)

	// ListByUserID はユーザーに関連するトランザクション一覧を取得
	ListByUserID(userID uuid.UUID, offset, limit int) ([]*Transaction, error)

	// ListAll は全トランザクション一覧を取得（管理者用）
	ListAll(offset, limit int) ([]*Transaction, error)

	// Update はトランザクションを更新
	Update(tx interface{}, transaction *Transaction) error

	// CountByUserID はユーザーのトランザクション総数を取得
	CountByUserID(userID uuid.UUID) (int64, error)
}

// IdempotencyKey は冪等性キーエンティティ
type IdempotencyKey struct {
	Key           string
	UserID        uuid.UUID
	TransactionID *uuid.UUID
	Status        string // processing, completed, failed
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

// NewIdempotencyKey は新しい冪等性キーを作成
func NewIdempotencyKey(key string, userID uuid.UUID) *IdempotencyKey {
	return &IdempotencyKey{
		Key:       key,
		UserID:    userID,
		Status:    "processing",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

// IdempotencyKeyRepository は冪等性キーのリポジトリインターフェース
type IdempotencyKeyRepository interface {
	// Create は新しい冪等性キーを作成（既存の場合はエラー）
	Create(key *IdempotencyKey) error

	// FindByKey はキーで冪等性キーを検索
	FindByKey(key string) (*IdempotencyKey, error)

	// Update は冪等性キーを更新
	Update(key *IdempotencyKey) error

	// DeleteExpired は期限切れの冪等性キーを削除
	DeleteExpired() error
}

// ヘルパー関数
func ptrTime(t time.Time) *time.Time {
	return &t
}
