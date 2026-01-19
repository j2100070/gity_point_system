package persistence

import (
	"errors"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FriendshipModel はGORM用の友達関係モデル
type FriendshipModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RequesterID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:unique_friendship,priority:1"`
	AddresseeID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:unique_friendship,priority:2"`
	Status      string    `gorm:"type:varchar(50);not null;default:'pending';index"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

// TableName はテーブル名を指定
func (FriendshipModel) TableName() string {
	return "friendships"
}

// ToDomain はドメインモデルに変換
func (f *FriendshipModel) ToDomain() *domain.Friendship {
	return &domain.Friendship{
		ID:          f.ID,
		RequesterID: f.RequesterID,
		AddresseeID: f.AddresseeID,
		Status:      domain.FriendshipStatus(f.Status),
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (f *FriendshipModel) FromDomain(friendship *domain.Friendship) {
	f.ID = friendship.ID
	f.RequesterID = friendship.RequesterID
	f.AddresseeID = friendship.AddresseeID
	f.Status = string(friendship.Status)
	f.CreatedAt = friendship.CreatedAt
	f.UpdatedAt = friendship.UpdatedAt
}

// FriendshipRepositoryImpl はFriendshipRepositoryの実装
type FriendshipRepositoryImpl struct {
	db *gorm.DB
}

// NewFriendshipRepository は新しいFriendshipRepositoryを作成
func NewFriendshipRepository(db *gorm.DB) domain.FriendshipRepository {
	return &FriendshipRepositoryImpl{db: db}
}

// Create は新しい友達申請を作成
func (r *FriendshipRepositoryImpl) Create(friendship *domain.Friendship) error {
	model := &FriendshipModel{}
	model.FromDomain(friendship)

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	*friendship = *model.ToDomain()
	return nil
}

// FindByID はIDで友達関係を検索
func (r *FriendshipRepositoryImpl) FindByID(id uuid.UUID) (*domain.Friendship, error) {
	var model FriendshipModel

	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("friendship not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// FindByUsers は2人のユーザー間の友達関係を検索
func (r *FriendshipRepositoryImpl) FindByUsers(userID1, userID2 uuid.UUID) (*domain.Friendship, error) {
	var model FriendshipModel

	err := r.db.Where(
		"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		userID1, userID2, userID2, userID1,
	).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("friendship not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// ListFriends は承認済みの友達一覧を取得
func (r *FriendshipRepositoryImpl) ListFriends(userID uuid.UUID, offset, limit int) ([]*domain.Friendship, error) {
	var models []FriendshipModel

	err := r.db.Where(
		"(requester_id = ? OR addressee_id = ?) AND status = ?",
		userID, userID, "accepted",
	).Offset(offset).
		Limit(limit).
		Order("updated_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	friendships := make([]*domain.Friendship, len(models))
	for i, model := range models {
		friendships[i] = model.ToDomain()
	}

	return friendships, nil
}

// ListPendingRequests は保留中の友達申請一覧を取得
func (r *FriendshipRepositoryImpl) ListPendingRequests(userID uuid.UUID, offset, limit int) ([]*domain.Friendship, error) {
	var models []FriendshipModel

	err := r.db.Where("addressee_id = ? AND status = ?", userID, "pending").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	friendships := make([]*domain.Friendship, len(models))
	for i, model := range models {
		friendships[i] = model.ToDomain()
	}

	return friendships, nil
}

// Update は友達関係を更新
func (r *FriendshipRepositoryImpl) Update(friendship *domain.Friendship) error {
	model := &FriendshipModel{}
	model.FromDomain(friendship)

	return r.db.Save(model).Error
}

// Delete は友達関係を削除
func (r *FriendshipRepositoryImpl) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&FriendshipModel{}).Error
}

// AreFriends は2人のユーザーが友達かどうかを確認
func (r *FriendshipRepositoryImpl) AreFriends(userID1, userID2 uuid.UUID) (bool, error) {
	var count int64

	err := r.db.Model(&FriendshipModel{}).Where(
		"((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)) AND status = ?",
		userID1, userID2, userID2, userID1, "accepted",
	).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
