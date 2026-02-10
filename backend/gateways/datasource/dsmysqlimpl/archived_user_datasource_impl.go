package dsmysqlimpl

import (
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArchivedUserModel はGORM用のアーカイブユーザーモデル
type ArchivedUserModel struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key"`
	Username          string     `gorm:"type:varchar(255);not null"`
	Email             string     `gorm:"type:varchar(255);not null"`
	PasswordHash      string     `gorm:"type:varchar(255);not null"`
	DisplayName       string     `gorm:"type:varchar(255);not null"`
	Balance           int64      `gorm:"not null"`
	Role              string     `gorm:"type:varchar(50);not null"`
	AvatarURL         *string    `gorm:"type:varchar(500)"`
	AvatarType        string     `gorm:"type:varchar(50)"`
	EmailVerified     bool       `gorm:"not null"`
	EmailVerifiedAt   *time.Time
	ArchivedAt        time.Time `gorm:"not null;default:now()"`
	ArchivedBy        *uuid.UUID
	DeletionReason    *string `gorm:"type:text"`
	OriginalCreatedAt time.Time `gorm:"not null"`
	OriginalUpdatedAt time.Time `gorm:"not null"`
}

// TableName はテーブル名を指定
func (ArchivedUserModel) TableName() string {
	return "archived_users"
}

// ToDomain はドメインモデルに変換
func (m *ArchivedUserModel) ToDomain() *entities.ArchivedUser {
	return &entities.ArchivedUser{
		ID:                m.ID,
		Username:          m.Username,
		Email:             m.Email,
		PasswordHash:      m.PasswordHash,
		DisplayName:       m.DisplayName,
		Balance:           m.Balance,
		Role:              entities.UserRole(m.Role),
		AvatarURL:         m.AvatarURL,
		AvatarType:        entities.AvatarType(m.AvatarType),
		EmailVerified:     m.EmailVerified,
		EmailVerifiedAt:   m.EmailVerifiedAt,
		ArchivedAt:        m.ArchivedAt,
		ArchivedBy:        m.ArchivedBy,
		DeletionReason:    m.DeletionReason,
		OriginalCreatedAt: m.OriginalCreatedAt,
		OriginalUpdatedAt: m.OriginalUpdatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (m *ArchivedUserModel) FromDomain(au *entities.ArchivedUser) {
	m.ID = au.ID
	m.Username = au.Username
	m.Email = au.Email
	m.PasswordHash = au.PasswordHash
	m.DisplayName = au.DisplayName
	m.Balance = au.Balance
	m.Role = string(au.Role)
	m.AvatarURL = au.AvatarURL
	m.AvatarType = string(au.AvatarType)
	m.EmailVerified = au.EmailVerified
	m.EmailVerifiedAt = au.EmailVerifiedAt
	m.ArchivedAt = au.ArchivedAt
	m.ArchivedBy = au.ArchivedBy
	m.DeletionReason = au.DeletionReason
	m.OriginalCreatedAt = au.OriginalCreatedAt
	m.OriginalUpdatedAt = au.OriginalUpdatedAt
}

// ArchivedUserDataSourceImpl はArchivedUserDataSourceの実装
type ArchivedUserDataSourceImpl struct {
	db inframysql.DB
}

// NewArchivedUserDataSource は新しいArchivedUserDataSourceを作成
func NewArchivedUserDataSource(db inframysql.DB) *ArchivedUserDataSourceImpl {
	return &ArchivedUserDataSourceImpl{db: db}
}

// Insert は新しいアーカイブユーザーを挿入
func (ds *ArchivedUserDataSourceImpl) Insert(archivedUser *entities.ArchivedUser) error {
	model := &ArchivedUserModel{}
	model.FromDomain(archivedUser)

	return ds.db.GetDB().Create(model).Error
}

// Select はIDでアーカイブユーザーを検索
func (ds *ArchivedUserDataSourceImpl) Select(id uuid.UUID) (*entities.ArchivedUser, error) {
	var model ArchivedUserModel

	err := ds.db.GetDB().Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("archived user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByUsername はユーザー名でアーカイブユーザーを検索
func (ds *ArchivedUserDataSourceImpl) SelectByUsername(username string) (*entities.ArchivedUser, error) {
	var model ArchivedUserModel

	err := ds.db.GetDB().Where("username = ?", username).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("archived user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectList はアーカイブユーザー一覧を取得
func (ds *ArchivedUserDataSourceImpl) SelectList(offset, limit int) ([]*entities.ArchivedUser, error) {
	var models []ArchivedUserModel

	err := ds.db.GetDB().
		Offset(offset).
		Limit(limit).
		Order("archived_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	users := make([]*entities.ArchivedUser, len(models))
	for i, model := range models {
		users[i] = model.ToDomain()
	}

	return users, nil
}

// Count はアーカイブユーザー総数を取得
func (ds *ArchivedUserDataSourceImpl) Count() (int64, error) {
	var count int64
	err := ds.db.GetDB().Model(&ArchivedUserModel{}).Count(&count).Error
	return count, err
}

// Delete はアーカイブユーザーを削除
func (ds *ArchivedUserDataSourceImpl) Delete(id uuid.UUID) error {
	return ds.db.GetDB().Delete(&ArchivedUserModel{}, "id = ?", id).Error
}

// Restore はアーカイブユーザーを復元（トランザクション内で使用）
func (ds *ArchivedUserDataSourceImpl) Restore(tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error {
	// トランザクションを*gorm.DBにキャスト
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New("invalid transaction type")
	}

	// アーカイブからユーザーを作成
	userModel := &UserModel{}
	userModel.FromDomain(user)

	if err := gormTx.Create(userModel).Error; err != nil {
		return err
	}

	// アーカイブから削除
	if err := gormTx.Delete(&ArchivedUserModel{}, "id = ?", archivedUser.ID).Error; err != nil {
		return err
	}

	return nil
}
