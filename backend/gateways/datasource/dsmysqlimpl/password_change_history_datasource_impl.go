package dsmysqlimpl

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
)

// PasswordChangeHistoryModel はGORM用のパスワード変更履歴モデル
type PasswordChangeHistoryModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	ChangedAt time.Time `gorm:"not null;default:now()"`
	IPAddress *string   `gorm:"type:inet"`
	UserAgent *string   `gorm:"type:text"`
}

// TableName はテーブル名を指定
func (PasswordChangeHistoryModel) TableName() string {
	return "password_change_history"
}

// ToDomain はドメインモデルに変換
func (m *PasswordChangeHistoryModel) ToDomain() *entities.PasswordChangeHistory {
	return &entities.PasswordChangeHistory{
		ID:        m.ID,
		UserID:    m.UserID,
		ChangedAt: m.ChangedAt,
		IPAddress: m.IPAddress,
		UserAgent: m.UserAgent,
	}
}

// FromDomain はドメインモデルから変換
func (m *PasswordChangeHistoryModel) FromDomain(history *entities.PasswordChangeHistory) {
	m.ID = history.ID
	m.UserID = history.UserID
	m.ChangedAt = history.ChangedAt
	m.IPAddress = history.IPAddress
	m.UserAgent = history.UserAgent
}

// PasswordChangeHistoryDataSourceImpl はPasswordChangeHistoryDataSourceの実装
type PasswordChangeHistoryDataSourceImpl struct {
	db inframysql.DB
}

// NewPasswordChangeHistoryDataSource は新しいPasswordChangeHistoryDataSourceを作成
func NewPasswordChangeHistoryDataSource(db inframysql.DB) *PasswordChangeHistoryDataSourceImpl {
	return &PasswordChangeHistoryDataSourceImpl{db: db}
}

// Insert は新しいパスワード変更履歴を挿入
func (ds *PasswordChangeHistoryDataSourceImpl) Insert(ctx context.Context, history *entities.PasswordChangeHistory) error {
	model := &PasswordChangeHistoryModel{}
	model.FromDomain(history)

	return inframysql.GetDB(ctx, ds.db.GetDB()).Create(model).Error
}

// SelectListByUserID はユーザーIDで履歴を取得
func (ds *PasswordChangeHistoryDataSourceImpl) SelectListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error) {
	var models []PasswordChangeHistoryModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("changed_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	histories := make([]*entities.PasswordChangeHistory, len(models))
	for i, model := range models {
		histories[i] = model.ToDomain()
	}

	return histories, nil
}

// CountByUserID はユーザーIDで履歴数を取得
func (ds *PasswordChangeHistoryDataSourceImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := inframysql.GetDB(ctx, ds.db.GetDB()).Model(&PasswordChangeHistoryModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
