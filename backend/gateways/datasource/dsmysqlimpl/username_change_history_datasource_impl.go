package dsmysqlimpl

import (
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
)

// UsernameChangeHistoryModel はGORM用のユーザー名変更履歴モデル
type UsernameChangeHistoryModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	OldUsername string     `gorm:"type:varchar(255);not null"`
	NewUsername string     `gorm:"type:varchar(255);not null"`
	ChangedAt   time.Time  `gorm:"not null;default:now()"`
	ChangedBy   *uuid.UUID `gorm:"type:uuid"`
	IPAddress   *string    `gorm:"type:inet"`
}

// TableName はテーブル名を指定
func (UsernameChangeHistoryModel) TableName() string {
	return "username_change_history"
}

// ToDomain はドメインモデルに変換
func (m *UsernameChangeHistoryModel) ToDomain() *entities.UsernameChangeHistory {
	return &entities.UsernameChangeHistory{
		ID:          m.ID,
		UserID:      m.UserID,
		OldUsername: m.OldUsername,
		NewUsername: m.NewUsername,
		ChangedAt:   m.ChangedAt,
		ChangedBy:   m.ChangedBy,
		IPAddress:   m.IPAddress,
	}
}

// FromDomain はドメインモデルから変換
func (m *UsernameChangeHistoryModel) FromDomain(history *entities.UsernameChangeHistory) {
	m.ID = history.ID
	m.UserID = history.UserID
	m.OldUsername = history.OldUsername
	m.NewUsername = history.NewUsername
	m.ChangedAt = history.ChangedAt
	m.ChangedBy = history.ChangedBy
	m.IPAddress = history.IPAddress
}

// UsernameChangeHistoryDataSourceImpl はUsernameChangeHistoryDataSourceの実装
type UsernameChangeHistoryDataSourceImpl struct {
	db inframysql.DB
}

// NewUsernameChangeHistoryDataSource は新しいUsernameChangeHistoryDataSourceを作成
func NewUsernameChangeHistoryDataSource(db inframysql.DB) *UsernameChangeHistoryDataSourceImpl {
	return &UsernameChangeHistoryDataSourceImpl{db: db}
}

// Insert は新しいユーザー名変更履歴を挿入
func (ds *UsernameChangeHistoryDataSourceImpl) Insert(history *entities.UsernameChangeHistory) error {
	model := &UsernameChangeHistoryModel{}
	model.FromDomain(history)

	return ds.db.GetDB().Create(model).Error
}

// SelectListByUserID はユーザーIDで履歴を取得
func (ds *UsernameChangeHistoryDataSourceImpl) SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error) {
	var models []UsernameChangeHistoryModel

	err := ds.db.GetDB().
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("changed_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	histories := make([]*entities.UsernameChangeHistory, len(models))
	for i, model := range models {
		histories[i] = model.ToDomain()
	}

	return histories, nil
}

// CountByUserID はユーザーIDで履歴数を取得
func (ds *UsernameChangeHistoryDataSourceImpl) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := ds.db.GetDB().Model(&UsernameChangeHistoryModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
