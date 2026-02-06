package dsmysqlimpl

import (
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserModel はGORM用のユーザーモデル
type UserModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	DisplayName  string     `gorm:"type:varchar(255);not null"`
	Balance      int64      `gorm:"not null;default:0;check:balance >= 0"`
	Role         string     `gorm:"type:varchar(50);not null;default:'user'"`
	Version      int        `gorm:"not null;default:1"`
	IsActive     bool       `gorm:"not null;default:true"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()"`
	DeletedAt    *time.Time `gorm:"index"`
}

// TableName はテーブル名を指定
func (UserModel) TableName() string {
	return "users"
}

// ToDomain はドメインモデルに変換
func (u *UserModel) ToDomain() *entities.User {
	return &entities.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		Balance:      u.Balance,
		Role:         entities.UserRole(u.Role),
		Version:      u.Version,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		DeletedAt:    u.DeletedAt,
	}
}

// FromDomain はドメインモデルから変換
func (u *UserModel) FromDomain(user *entities.User) {
	u.ID = user.ID
	u.Username = user.Username
	u.Email = user.Email
	u.PasswordHash = user.PasswordHash
	u.DisplayName = user.DisplayName
	u.Balance = user.Balance
	u.Role = string(user.Role)
	u.Version = user.Version
	u.IsActive = user.IsActive
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
	u.DeletedAt = user.DeletedAt
}

// UserDataSourceImpl はUserDataSourceの実装
// Infraを活用し、Repositoryが要求するデータの取得、永続化を達成
type UserDataSourceImpl struct {
	db inframysql.DB
}

// NewUserDataSource は新しいUserDataSourceを作成
func NewUserDataSource(db inframysql.DB) dsmysql.UserDataSource {
	return &UserDataSourceImpl{db: db}
}

// Insert は新しいユーザーを挿入
func (ds *UserDataSourceImpl) Insert(user *entities.User) error {
	model := &UserModel{}
	model.FromDomain(user)

	if err := ds.db.GetDB().Create(model).Error; err != nil {
		return err
	}

	*user = *model.ToDomain()
	return nil
}

// Select はIDでユーザーを検索
func (ds *UserDataSourceImpl) Select(id uuid.UUID) (*entities.User, error) {
	var model UserModel

	err := ds.db.GetDB().Where("id = ? AND deleted_at IS NULL", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByUsername はユーザー名でユーザーを検索
func (ds *UserDataSourceImpl) SelectByUsername(username string) (*entities.User, error) {
	var model UserModel

	err := ds.db.GetDB().Where("username = ? AND deleted_at IS NULL", username).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByEmail はメールアドレスでユーザーを検索
func (ds *UserDataSourceImpl) SelectByEmail(email string) (*entities.User, error) {
	var model UserModel

	err := ds.db.GetDB().Where("email = ? AND deleted_at IS NULL", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はユーザー情報を更新（楽観的ロック対応）
func (ds *UserDataSourceImpl) Update(user *entities.User) (bool, error) {
	model := &UserModel{}
	model.FromDomain(user)

	// 楽観的ロック: versionが一致する場合のみ更新
	result := ds.db.GetDB().Model(&UserModel{}).
		Where("id = ? AND version = ? AND deleted_at IS NULL", user.ID, user.Version-1).
		Updates(map[string]interface{}{
			"username":      model.Username,
			"email":         model.Email,
			"password_hash": model.PasswordHash,
			"display_name":  model.DisplayName,
			"balance":       model.Balance,
			"role":          model.Role,
			"version":       model.Version,
			"is_active":     model.IsActive,
			"updated_at":    time.Now(),
		})

	if result.Error != nil {
		return false, result.Error
	}

	// 更新された行数が0の場合は楽観的ロック失敗
	if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

// UpdateBalanceWithLock は残高を更新（悲観的ロック: SELECT FOR UPDATE）
func (ds *UserDataSourceImpl) UpdateBalanceWithLock(tx interface{}, userID uuid.UUID, amount int64, isDeduct bool) error {
	db := ds.db.GetDB()
	if tx != nil {
		db = tx.(*gorm.DB)
	}

	var model UserModel

	// SELECT FOR UPDATE で行ロック
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// 残高チェック（減算の場合）
	if isDeduct && model.Balance < amount {
		return errors.New("insufficient balance")
	}

	// 残高更新
	newBalance := model.Balance
	if isDeduct {
		newBalance -= amount
	} else {
		newBalance += amount
	}

	// 負の値チェック
	if newBalance < 0 {
		return errors.New("balance cannot be negative")
	}

	// 更新実行
	err = db.Model(&model).Updates(map[string]interface{}{
		"balance":    newBalance,
		"version":    gorm.Expr("version + 1"),
		"updated_at": time.Now(),
	}).Error

	return err
}

// UpdateBalancesWithLock は複数ユーザーの残高を一括更新（悲観的ロック、デッドロック回避）
// デッドロック回避のために、常にUUID順（小さい順）にSELECT FOR UPDATEを実行します
func (ds *UserDataSourceImpl) UpdateBalancesWithLock(tx interface{}, updates []dsmysql.BalanceUpdate) error {
	db := ds.db.GetDB()
	if tx != nil {
		db = tx.(*gorm.DB)
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	// ID順にソート（デッドロック回避のため）
	sortedUpdates := make([]dsmysql.BalanceUpdate, len(updates))
	copy(sortedUpdates, updates)

	// UUID文字列で比較してソート
	for i := 0; i < len(sortedUpdates)-1; i++ {
		for j := i + 1; j < len(sortedUpdates); j++ {
			if sortedUpdates[j].UserID.String() < sortedUpdates[i].UserID.String() {
				sortedUpdates[i], sortedUpdates[j] = sortedUpdates[j], sortedUpdates[i]
			}
		}
	}

	// ソート順にロックを取得し、残高を更新
	for _, update := range sortedUpdates {
		var model UserModel

		// SELECT FOR UPDATE で行ロック
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", update.UserID).
			First(&model).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		// 残高チェック（減算の場合）
		if update.IsDeduct && model.Balance < update.Amount {
			return errors.New("insufficient balance")
		}

		// 残高更新
		newBalance := model.Balance
		if update.IsDeduct {
			newBalance -= update.Amount
		} else {
			newBalance += update.Amount
		}

		// 負の値チェック
		if newBalance < 0 {
			return errors.New("balance cannot be negative")
		}

		// 更新実行
		err = db.Model(&model).Updates(map[string]interface{}{
			"balance":    newBalance,
			"version":    gorm.Expr("version + 1"),
			"updated_at": time.Now(),
		}).Error

		if err != nil {
			return err
		}
	}

	return nil
}

// SelectList はユーザー一覧を取得
func (ds *UserDataSourceImpl) SelectList(offset, limit int) ([]*entities.User, error) {
	var models []UserModel

	err := ds.db.GetDB().Where("deleted_at IS NULL").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	users := make([]*entities.User, len(models))
	for i, model := range models {
		users[i] = model.ToDomain()
	}

	return users, nil
}

// Count はユーザー総数を取得
func (ds *UserDataSourceImpl) Count() (int64, error) {
	var count int64
	err := ds.db.GetDB().Model(&UserModel{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

// Delete はユーザーを論理削除
func (ds *UserDataSourceImpl) Delete(id uuid.UUID) error {
	now := time.Now()
	return ds.db.GetDB().Model(&UserModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
}
