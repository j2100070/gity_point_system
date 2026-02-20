package friendship

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// RepositoryImpl はFriendshipRepositoryの実装
type RepositoryImpl struct {
	friendshipDS dsmysql.FriendshipDataSource
	logger       entities.Logger
}

// NewFriendshipRepository は新しいFriendshipRepositoryを作成
func NewFriendshipRepository(
	friendshipDS dsmysql.FriendshipDataSource,
	logger entities.Logger,
) repository.FriendshipRepository {
	return &RepositoryImpl{
		friendshipDS: friendshipDS,
		logger:       logger,
	}
}

// Create は新しい友達申請を作成
func (r *RepositoryImpl) Create(ctx context.Context, friendship *entities.Friendship) error {
	r.logger.Debug("Creating friendship request",
		entities.NewField("requester_id", friendship.RequesterID),
		entities.NewField("addressee_id", friendship.AddresseeID))
	return r.friendshipDS.Insert(ctx, friendship)
}

// Read はIDで友達関係を検索
func (r *RepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.Friendship, error) {
	return r.friendshipDS.Select(ctx, id)
}

// ReadByUsers は2人のユーザー間の友達関係を検索
func (r *RepositoryImpl) ReadByUsers(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Friendship, error) {
	return r.friendshipDS.SelectByUsers(ctx, userID1, userID2)
}

// ReadListFriends は承認済みの友達一覧を取得
func (r *RepositoryImpl) ReadListFriends(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return r.friendshipDS.SelectListFriends(ctx, userID, offset, limit)
}

// ReadListPendingRequests は保留中の友達申請一覧を取得
func (r *RepositoryImpl) ReadListPendingRequests(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return r.friendshipDS.SelectListPendingRequests(ctx, userID, offset, limit)
}

// Update は友達関係を更新
func (r *RepositoryImpl) Update(ctx context.Context, friendship *entities.Friendship) error {
	r.logger.Debug("Updating friendship", entities.NewField("friendship_id", friendship.ID))
	return r.friendshipDS.Update(ctx, friendship)
}

// Delete は友達関係を削除
func (r *RepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting friendship", entities.NewField("friendship_id", id))
	return r.friendshipDS.Delete(ctx, id)
}

// ArchiveAndDelete は友達関係をアーカイブしてから削除
func (r *RepositoryImpl) ArchiveAndDelete(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	r.logger.Debug("Archiving and deleting friendship",
		entities.NewField("friendship_id", id),
		entities.NewField("archived_by", archivedBy))
	return r.friendshipDS.ArchiveAndDelete(ctx, id, archivedBy)
}

// CheckAreFriends は2人のユーザーが友達かどうかを確認
func (r *RepositoryImpl) CheckAreFriends(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	return r.friendshipDS.CheckAreFriends(ctx, userID1, userID2)
}

// ReadListFriendsWithUsers は承認済みの友達一覧をユーザー情報付きで取得（JOIN）
func (r *RepositoryImpl) ReadListFriendsWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.FriendshipWithUser, error) {
	return r.friendshipDS.SelectListFriendsWithUsers(ctx, userID, offset, limit)
}

// ReadListPendingRequestsWithUsers は保留中の友達申請一覧をユーザー情報付きで取得（JOIN）
func (r *RepositoryImpl) ReadListPendingRequestsWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.FriendshipWithUser, error) {
	return r.friendshipDS.SelectListPendingRequestsWithUsers(ctx, userID, offset, limit)
}

// CountPendingRequests は保留中の友達申請件数を取得
func (r *RepositoryImpl) CountPendingRequests(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.friendshipDS.CountPendingRequests(ctx, userID)
}
