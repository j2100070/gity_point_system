package dsmysqlimpl

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FriendshipModel はGORM用の友達関係モデル
type FriendshipModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RequesterID uuid.UUID `gorm:"type:uuid;not null;index"`
	AddresseeID uuid.UUID `gorm:"type:uuid;not null;index"`
	Status      string    `gorm:"type:varchar(50);not null;index"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

// TableName はテーブル名を指定
func (FriendshipModel) TableName() string {
	return "friendships"
}

// ToDomain はドメインモデルに変換
func (f *FriendshipModel) ToDomain() *entities.Friendship {
	return &entities.Friendship{
		ID:          f.ID,
		RequesterID: f.RequesterID,
		AddresseeID: f.AddresseeID,
		Status:      entities.FriendshipStatus(f.Status),
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (f *FriendshipModel) FromDomain(friendship *entities.Friendship) {
	f.ID = friendship.ID
	f.RequesterID = friendship.RequesterID
	f.AddresseeID = friendship.AddresseeID
	f.Status = string(friendship.Status)
	f.CreatedAt = friendship.CreatedAt
	f.UpdatedAt = friendship.UpdatedAt
}

// FriendshipDataSourceImpl はFriendshipDataSourceの実装
type FriendshipDataSourceImpl struct {
	db inframysql.DB
}

// NewFriendshipDataSource は新しいFriendshipDataSourceを作成
func NewFriendshipDataSource(db inframysql.DB) dsmysql.FriendshipDataSource {
	return &FriendshipDataSourceImpl{db: db}
}

// Insert は新しい友達申請を挿入
func (ds *FriendshipDataSourceImpl) Insert(ctx context.Context, friendship *entities.Friendship) error {
	model := &FriendshipModel{}
	model.FromDomain(friendship)

	if err := inframysql.GetDB(ctx, ds.db.GetDB()).Create(model).Error; err != nil {
		return err
	}

	*friendship = *model.ToDomain()
	return nil
}

// Select はIDで友達関係を検索
func (ds *FriendshipDataSourceImpl) Select(ctx context.Context, id uuid.UUID) (*entities.Friendship, error) {
	var model FriendshipModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("friendship not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByUsers は2人のユーザー間の友達関係を検索
func (ds *FriendshipDataSourceImpl) SelectByUsers(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Friendship, error) {
	var model FriendshipModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
			userID1, userID2, userID2, userID1).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("friendship not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectListFriends は承認済みの友達一覧を取得
func (ds *FriendshipDataSourceImpl) SelectListFriends(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	var models []FriendshipModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?",
			userID, userID, "accepted").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	friendships := make([]*entities.Friendship, len(models))
	for i, model := range models {
		friendships[i] = model.ToDomain()
	}

	return friendships, nil
}

// SelectListPendingRequests は保留中の友達申請一覧を取得
func (ds *FriendshipDataSourceImpl) SelectListPendingRequests(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	var models []FriendshipModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("addressee_id = ? AND status = ?", userID, "pending").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	friendships := make([]*entities.Friendship, len(models))
	for i, model := range models {
		friendships[i] = model.ToDomain()
	}

	return friendships, nil
}

// Update は友達関係を更新
func (ds *FriendshipDataSourceImpl) Update(ctx context.Context, friendship *entities.Friendship) error {
	model := &FriendshipModel{}
	model.FromDomain(friendship)

	return inframysql.GetDB(ctx, ds.db.GetDB()).Model(&FriendshipModel{}).
		Where("id = ?", friendship.ID).
		Updates(map[string]interface{}{
			"status":     model.Status,
			"updated_at": time.Now(),
		}).Error
}

// Delete は友達関係を削除
func (ds *FriendshipDataSourceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return inframysql.GetDB(ctx, ds.db.GetDB()).Where("id = ?", id).Delete(&FriendshipModel{}).Error
}

// FriendshipArchiveModel はGORM用のアーカイブモデル
type FriendshipArchiveModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	RequesterID uuid.UUID `gorm:"type:uuid;not null"`
	AddresseeID uuid.UUID `gorm:"type:uuid;not null"`
	Status      string    `gorm:"type:varchar(50);not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	ArchivedAt  time.Time `gorm:"not null;default:now()"`
	ArchivedBy  uuid.UUID `gorm:"type:uuid"`
}

func (FriendshipArchiveModel) TableName() string {
	return "friendships_archive"
}

// ArchiveAndDelete は友達関係をアーカイブしてから削除
func (ds *FriendshipDataSourceImpl) ArchiveAndDelete(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	return inframysql.GetDB(ctx, ds.db.GetDB()).Transaction(func(tx *gorm.DB) error {
		// 1. 元レコードを取得
		var model FriendshipModel
		if err := tx.Where("id = ?", id).First(&model).Error; err != nil {
			return err
		}

		// 2. アーカイブテーブルにINSERT
		archive := FriendshipArchiveModel{
			ID:          model.ID,
			RequesterID: model.RequesterID,
			AddresseeID: model.AddresseeID,
			Status:      model.Status,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
			ArchivedAt:  time.Now(),
			ArchivedBy:  archivedBy,
		}
		if err := tx.Create(&archive).Error; err != nil {
			return err
		}

		// 3. 元テーブルから削除
		return tx.Where("id = ?", id).Delete(&FriendshipModel{}).Error
	})
}

// CheckAreFriends は2人のユーザーが友達かどうかを確認
func (ds *FriendshipDataSourceImpl) CheckAreFriends(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	var count int64

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Model(&FriendshipModel{}).
		Where("((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)) AND status = ?",
			userID1, userID2, userID2, userID1, "accepted").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
