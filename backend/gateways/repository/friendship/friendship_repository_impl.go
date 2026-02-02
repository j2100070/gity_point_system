package friendship

import (
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
func (r *RepositoryImpl) Create(friendship *entities.Friendship) error {
	r.logger.Debug("Creating friendship request",
		entities.NewField("requester_id", friendship.RequesterID),
		entities.NewField("addressee_id", friendship.AddresseeID))
	return r.friendshipDS.Insert(friendship)
}

// Read はIDで友達関係を検索
func (r *RepositoryImpl) Read(id uuid.UUID) (*entities.Friendship, error) {
	return r.friendshipDS.Select(id)
}

// ReadByUsers は2人のユーザー間の友達関係を検索
func (r *RepositoryImpl) ReadByUsers(userID1, userID2 uuid.UUID) (*entities.Friendship, error) {
	return r.friendshipDS.SelectByUsers(userID1, userID2)
}

// ReadListFriends は承認済みの友達一覧を取得
func (r *RepositoryImpl) ReadListFriends(userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return r.friendshipDS.SelectListFriends(userID, offset, limit)
}

// ReadListPendingRequests は保留中の友達申請一覧を取得
func (r *RepositoryImpl) ReadListPendingRequests(userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return r.friendshipDS.SelectListPendingRequests(userID, offset, limit)
}

// Update は友達関係を更新
func (r *RepositoryImpl) Update(friendship *entities.Friendship) error {
	r.logger.Debug("Updating friendship", entities.NewField("friendship_id", friendship.ID))
	return r.friendshipDS.Update(friendship)
}

// Delete は友達関係を削除
func (r *RepositoryImpl) Delete(id uuid.UUID) error {
	r.logger.Debug("Deleting friendship", entities.NewField("friendship_id", id))
	return r.friendshipDS.Delete(id)
}

// CheckAreFriends は2人のユーザーが友達かどうかを確認
func (r *RepositoryImpl) CheckAreFriends(userID1, userID2 uuid.UUID) (bool, error) {
	return r.friendshipDS.CheckAreFriends(userID1, userID2)
}
