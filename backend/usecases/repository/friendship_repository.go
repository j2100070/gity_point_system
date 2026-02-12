package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// FriendshipRepository は友達関係のリポジトリインターフェース
type FriendshipRepository interface {
	// Create は新しい友達申請を作成
	Create(ctx context.Context, friendship *entities.Friendship) error

	// Read はIDで友達関係を検索
	Read(ctx context.Context, id uuid.UUID) (*entities.Friendship, error)

	// ReadByUsers は2人のユーザー間の友達関係を検索
	ReadByUsers(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Friendship, error)

	// ReadListFriends は承認済みの友達一覧を取得
	ReadListFriends(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error)

	// ReadListPendingRequests は保留中の友達申請一覧を取得
	ReadListPendingRequests(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error)

	// Update は友達関係を更新
	Update(ctx context.Context, friendship *entities.Friendship) error

	// Delete は友達関係を削除
	Delete(ctx context.Context, id uuid.UUID) error

	// ArchiveAndDelete は友達関係をアーカイブテーブルに移動してから削除
	ArchiveAndDelete(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error

	// CheckAreFriends は2人のユーザーが友達かどうかを確認
	CheckAreFriends(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error)
}
