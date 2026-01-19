package persistence

import (
	"errors"
	"time"

	"github.com/gity/point-system/internal/domain"
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
func (u *UserModel) ToDomain() *domain.User {
	return &domain.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		Balance:      u.Balance,
		Role:         domain.UserRole(u.Role),
		Version:      u.Version,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		DeletedAt:    u.DeletedAt,
	}
}

// FromDomain はドメインモデルから変換
func (u *UserModel) FromDomain(user *domain.User) {
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

// UserRepositoryImpl はUserRepositoryの実装
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository は新しいUserRepositoryを作成
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create は新しいユーザーを作成
func (r *UserRepositoryImpl) Create(user *domain.User) error {
	model := &UserModel{}
	model.FromDomain(user)

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	*user = *model.ToDomain()
	return nil
}

// FindByID はIDでユーザーを検索
func (r *UserRepositoryImpl) FindByID(id uuid.UUID) (*domain.User, error) {
	var model UserModel

	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// FindByUsername はユーザー名でユーザーを検索
func (r *UserRepositoryImpl) FindByUsername(username string) (*domain.User, error) {
	var model UserModel

	err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// FindByEmail はメールアドレスでユーザーを検索
func (r *UserRepositoryImpl) FindByEmail(email string) (*domain.User, error) {
	var model UserModel

	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はユーザー情報を更新（楽観的ロック対応）
func (r *UserRepositoryImpl) Update(user *domain.User) (bool, error) {
	model := &UserModel{}
	model.FromDomain(user)

	// 楽観的ロック: versionが一致する場合のみ更新
	result := r.db.Model(&UserModel{}).
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
// CRITICAL: トランザクション内で使用する必要がある
func (r *UserRepositoryImpl) UpdateBalanceWithLock(tx interface{}, userID uuid.UUID, amount int64, isDeduct bool) error {
	db := r.db
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

// List はユーザー一覧を取得（ページネーション対応）
func (r *UserRepositoryImpl) List(offset, limit int) ([]*domain.User, error) {
	var models []UserModel

	err := r.db.Where("deleted_at IS NULL").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(models))
	for i, model := range models {
		users[i] = model.ToDomain()
	}

	return users, nil
}

// Count はユーザー総数を取得
func (r *UserRepositoryImpl) Count() (int64, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

// SoftDelete はユーザーを論理削除
func (r *UserRepositoryImpl) SoftDelete(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&UserModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
}
