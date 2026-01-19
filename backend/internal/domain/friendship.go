package domain

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

// FriendshipRepository は友達関係のリポジトリインターフェース
type FriendshipRepository interface {
	// Create は新しい友達申請を作成
	Create(friendship *Friendship) error

	// FindByID はIDで友達関係を検索
	FindByID(id uuid.UUID) (*Friendship, error)

	// FindByUsers は2人のユーザー間の友達関係を検索
	FindByUsers(userID1, userID2 uuid.UUID) (*Friendship, error)

	// ListFriends は承認済みの友達一覧を取得
	ListFriends(userID uuid.UUID, offset, limit int) ([]*Friendship, error)

	// ListPendingRequests は保留中の友達申請一覧を取得
	ListPendingRequests(userID uuid.UUID, offset, limit int) ([]*Friendship, error)

	// Update は友達関係を更新
	Update(friendship *Friendship) error

	// Delete は友達関係を削除
	Delete(id uuid.UUID) error

	// AreFriends は2人のユーザーが友達かどうかを確認
	AreFriends(userID1, userID2 uuid.UUID) (bool, error)
}
