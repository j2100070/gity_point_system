package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// FriendshipDataSource はMySQLの友達関係データソースインターフェース
type FriendshipDataSource interface {
	// Insert は新しい友達申請を挿入
	Insert(friendship *entities.Friendship) error

	// Select はIDで友達関係を検索
	Select(id uuid.UUID) (*entities.Friendship, error)

	// SelectByUsers は2人のユーザー間の友達関係を検索
	SelectByUsers(userID1, userID2 uuid.UUID) (*entities.Friendship, error)

	// SelectListFriends は承認済みの友達一覧を取得
	SelectListFriends(userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error)

	// SelectListPendingRequests は保留中の友達申請一覧を取得
	SelectListPendingRequests(userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error)

	// Update は友達関係を更新
	Update(friendship *entities.Friendship) error

	// Delete は友達関係を削除
	Delete(id uuid.UUID) error

	// ArchiveAndDelete は友達関係をアーカイブテーブルに移動してから削除
	ArchiveAndDelete(id uuid.UUID, archivedBy uuid.UUID) error

	// CheckAreFriends は2人のユーザーが友達かどうかを確認
	CheckAreFriends(userID1, userID2 uuid.UUID) (bool, error)
}
