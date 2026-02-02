package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// FriendshipStatus は友達関係のステータス
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
	FriendshipStatusRejected FriendshipStatus = "rejected"
	FriendshipStatusBlocked  FriendshipStatus = "blocked"
)

// Friendship は友達関係エンティティ
type Friendship struct {
	ID          uuid.UUID
	RequesterID uuid.UUID        // 申請者
	AddresseeID uuid.UUID        // 受信者
	Status      FriendshipStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewFriendship は新しい友達申請を作成
func NewFriendship(requesterID, addresseeID uuid.UUID) (*Friendship, error) {
	if requesterID == addresseeID {
		return nil, errors.New("cannot send friend request to yourself")
	}

	return &Friendship{
		ID:          uuid.New(),
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      FriendshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// Accept は友達申請を承認
func (f *Friendship) Accept() error {
	if f.Status != FriendshipStatusPending {
		return errors.New("friendship is not in pending status")
	}
	f.Status = FriendshipStatusAccepted
	f.UpdatedAt = time.Now()
	return nil
}

// Reject は友達申請を拒否
func (f *Friendship) Reject() error {
	if f.Status != FriendshipStatusPending {
		return errors.New("friendship is not in pending status")
	}
	f.Status = FriendshipStatusRejected
	f.UpdatedAt = time.Now()
	return nil
}

// Block はユーザーをブロック
func (f *Friendship) Block() {
	f.Status = FriendshipStatusBlocked
	f.UpdatedAt = time.Now()
}

// IsAccepted は友達関係が承認済みかどうかを確認
func (f *Friendship) IsAccepted() bool {
	return f.Status == FriendshipStatusAccepted
}
