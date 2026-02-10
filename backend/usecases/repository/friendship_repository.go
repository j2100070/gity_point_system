package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// FriendshipRepository は友達関係のリポジトリインターフェース
type FriendshipRepository interface {
	// Create は新しい友達申請を作成
	Create(friendship *entities.Friendship) error

	// Read はIDで友達関係を検索
	Read(id uuid.UUID) (*entities.Friendship, error)

	// ReadByUsers は2人のユーザー間の友達関係を検索
	ReadByUsers(userID1, userID2 uuid.UUID) (*entities.Friendship, error)

	// ReadListFriends は承認済みの友達一覧を取得
	ReadListFriends(userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error)

	// ReadListPendingRequests は保留中の友達申請一覧を取得
	ReadListPendingRequests(userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error)

	// Update は友達関係を更新
	Update(friendship *entities.Friendship) error

	// Delete は友達関係を削除
	Delete(id uuid.UUID) error

	// ArchiveAndDelete は友達関係をアーカイブテーブルに移動してから削除
	ArchiveAndDelete(id uuid.UUID, archivedBy uuid.UUID) error

	// CheckAreFriends は2人のユーザーが友達かどうかを確認
	CheckAreFriends(userID1, userID2 uuid.UUID) (bool, error)
}
