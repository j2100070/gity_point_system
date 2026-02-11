package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TransferRequestStatus は送金リクエストの状態
type TransferRequestStatus string

const (
	TransferRequestStatusPending   TransferRequestStatus = "pending"   // 承認待ち
	TransferRequestStatusApproved  TransferRequestStatus = "approved"  // 承認済み
	TransferRequestStatusRejected  TransferRequestStatus = "rejected"  // 拒否
	TransferRequestStatusCancelled TransferRequestStatus = "cancelled" // キャンセル
	TransferRequestStatusExpired   TransferRequestStatus = "expired"   // 期限切れ
)

// TransferRequest は送金リクエストエンティティ
type TransferRequest struct {
	ID             uuid.UUID
	FromUserID     uuid.UUID  // 送信者
	ToUserID       uuid.UUID  // 受取人
	Amount         int64      // 送金額
	Message        string     // オプショナルメモ
	Status         TransferRequestStatus
	IdempotencyKey string     // 重複防止キー
	ExpiresAt      time.Time  // 有効期限（24時間）
	ApprovedAt     *time.Time // 承認日時
	RejectedAt     *time.Time // 拒否日時
	CancelledAt    *time.Time // キャンセル日時
	TransactionID  *uuid.UUID // 承認後に作成されるTransaction ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewTransferRequest は新しい送金リクエストを作成
func NewTransferRequest(fromUserID, toUserID uuid.UUID, amount int64, message, idempotencyKey string) (*TransferRequest, error) {
	if fromUserID == uuid.Nil {
		return nil, errors.New("from_user_id is required")
	}
	if toUserID == uuid.Nil {
		return nil, errors.New("to_user_id is required")
	}
	if fromUserID == toUserID {
		return nil, errors.New("cannot send to yourself")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency_key is required")
	}

	now := time.Now()
	return &TransferRequest{
		ID:             uuid.New(),
		FromUserID:     fromUserID,
		ToUserID:       toUserID,
		Amount:         amount,
		Message:        message,
		Status:         TransferRequestStatusPending,
		IdempotencyKey: idempotencyKey,
		ExpiresAt:      now.Add(24 * time.Hour), // 24時間有効
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// IsExpired はリクエストが期限切れかどうかを確認
func (tr *TransferRequest) IsExpired() bool {
	return time.Now().After(tr.ExpiresAt)
}

// IsPending はリクエストが承認待ちかどうかを確認
func (tr *TransferRequest) IsPending() bool {
	return tr.Status == TransferRequestStatusPending
}

// CanApprove は承認可能かどうかを確認
func (tr *TransferRequest) CanApprove() error {
	if tr.Status != TransferRequestStatusPending {
		return errors.New("request is not pending")
	}
	if tr.IsExpired() {
		return errors.New("request has expired")
	}
	return nil
}

// CanReject は拒否可能かどうかを確認
func (tr *TransferRequest) CanReject() error {
	if tr.Status != TransferRequestStatusPending {
		return errors.New("request is not pending")
	}
	return nil
}

// CanCancel はキャンセル可能かどうかを確認
func (tr *TransferRequest) CanCancel() error {
	if tr.Status != TransferRequestStatusPending {
		return errors.New("request is not pending")
	}
	return nil
}

// Approve はリクエストを承認
func (tr *TransferRequest) Approve(transactionID uuid.UUID) error {
	if err := tr.CanApprove(); err != nil {
		return err
	}

	now := time.Now()
	tr.Status = TransferRequestStatusApproved
	tr.ApprovedAt = &now
	tr.TransactionID = &transactionID
	tr.UpdatedAt = now
	return nil
}

// Reject はリクエストを拒否
func (tr *TransferRequest) Reject() error {
	if err := tr.CanReject(); err != nil {
		return err
	}

	now := time.Now()
	tr.Status = TransferRequestStatusRejected
	tr.RejectedAt = &now
	tr.UpdatedAt = now
	return nil
}

// Cancel はリクエストをキャンセル
func (tr *TransferRequest) Cancel() error {
	if err := tr.CanCancel(); err != nil {
		return err
	}

	now := time.Now()
	tr.Status = TransferRequestStatusCancelled
	tr.CancelledAt = &now
	tr.UpdatedAt = now
	return nil
}

// MarkAsExpired はリクエストを期限切れにマーク
func (tr *TransferRequest) MarkAsExpired() {
	if tr.Status == TransferRequestStatusPending && tr.IsExpired() {
		tr.Status = TransferRequestStatusExpired
		tr.UpdatedAt = time.Now()
	}
}
