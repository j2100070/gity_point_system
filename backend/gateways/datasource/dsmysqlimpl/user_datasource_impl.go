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
	"gorm.io/gorm/clause"
)

// UserModel はGORMのユーザーモデル
type UserModel struct {
	ID              string     `gorm:"column:id;primaryKey;type:char(36)"`
	Username        string     `gorm:"column:username;uniqueIndex;not null"`
	Email           string     `gorm:"column:email;uniqueIndex;not null"`
	PasswordHash    string     `gorm:"column:password_hash;not null"`
	DisplayName     string     `gorm:"column:display_name;not null"`
	FirstName       string     `gorm:"column:first_name;not null;default:''"`
	LastName        string     `gorm:"column:last_name;not null;default:''"`
	Balance         int64      `gorm:"column:balance;not null;default:0"`
	Role            string     `gorm:"column:role;not null;default:'user'"`
	Version         int        `gorm:"column:version;not null;default:1"`
	IsActive        bool       `gorm:"column:is_active;not null;default:true"`
	AvatarURL       *string    `gorm:"column:avatar_url"`
	AvatarType      string     `gorm:"column:avatar_type;not null;default:'generated'"`
	PersonalQRCode  string     `gorm:"column:personal_qr_code"`
	EmailVerified   bool       `gorm:"column:email_verified;not null;default:false"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName はテーブル名を指定
func (UserModel) TableName() string {
	return "users"
}

// ToDomain はドメインモデルに変換
func (m *UserModel) ToDomain() *entities.User {
	userID, _ := uuid.Parse(m.ID)
	return &entities.User{
		ID:              userID,
		Username:        m.Username,
		Email:           m.Email,
		PasswordHash:    m.PasswordHash,
		DisplayName:     m.DisplayName,
		FirstName:       m.FirstName,
		LastName:        m.LastName,
		Balance:         m.Balance,
		Role:            entities.UserRole(m.Role),
		Version:         m.Version,
		IsActive:        m.IsActive,
		AvatarURL:       m.AvatarURL,
		AvatarType:      entities.AvatarType(m.AvatarType),
		PersonalQRCode:  m.PersonalQRCode,
		EmailVerified:   m.EmailVerified,
		EmailVerifiedAt: m.EmailVerifiedAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (u *UserModel) FromDomain(user *entities.User) {
	u.ID = user.ID.String()
	u.Username = user.Username
	u.Email = user.Email
	u.PasswordHash = user.PasswordHash
	u.DisplayName = user.DisplayName
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.Balance = user.Balance
	u.Role = string(user.Role)
	u.Version = user.Version
	u.IsActive = user.IsActive
	u.AvatarURL = user.AvatarURL
	u.AvatarType = string(user.AvatarType)
	u.PersonalQRCode = user.PersonalQRCode
	u.EmailVerified = user.EmailVerified
	u.EmailVerifiedAt = user.EmailVerifiedAt
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
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
func (ds *UserDataSourceImpl) Insert(ctx context.Context, user *entities.User) error {
	model := &UserModel{}
	model.FromDomain(user)

	db := inframysql.GetDB(ctx, ds.db.GetDB())
	if err := db.Create(model).Error; err != nil {
		return err
	}

	*user = *model.ToDomain()
	return nil
}

// Select はIDでユーザーを検索
func (ds *UserDataSourceImpl) Select(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model UserModel

	err := db.Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByUsername はユーザー名でユーザーを検索
func (ds *UserDataSourceImpl) SelectByUsername(ctx context.Context, username string) (*entities.User, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model UserModel

	err := db.Where("username = ?", username).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByEmail はメールアドレスでユーザーを検索
func (ds *UserDataSourceImpl) SelectByEmail(ctx context.Context, email string) (*entities.User, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model UserModel

	err := db.Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はユーザー情報を更新（楽観的ロック対応）
func (ds *UserDataSourceImpl) Update(ctx context.Context, user *entities.User) (bool, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := &UserModel{}
	model.FromDomain(user)

	// 楽観的ロック: versionが一致する場合のみ更新
	result := db.Model(&UserModel{}).Where("id = ? AND version = ?", user.ID.String(), user.Version-1).
		Updates(map[string]interface{}{
			"username":          model.Username,
			"email":             model.Email,
			"password_hash":     model.PasswordHash,
			"display_name":      model.DisplayName,
			"first_name":        model.FirstName,
			"last_name":         model.LastName,
			"balance":           model.Balance,
			"role":              model.Role,
			"version":           model.Version,
			"is_active":         model.IsActive,
			"avatar_url":        model.AvatarURL,
			"avatar_type":       model.AvatarType,
			"email_verified":    model.EmailVerified,
			"email_verified_at": model.EmailVerifiedAt,
			"updated_at":        time.Now(),
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

// UpdatePartial は指定されたフィールドのみを更新（楽観的ロックなし）
func (ds *UserDataSourceImpl) UpdatePartial(ctx context.Context, userID uuid.UUID, fields map[string]interface{}) (bool, error) {
	// updated_atを自動追加
	fields["updated_at"] = time.Now()

	db := inframysql.GetDB(ctx, ds.db.GetDB())
	result := db.Model(&UserModel{}).
		Where("id = ?", userID).
		Updates(fields)

	if result.Error != nil {
		return false, result.Error
	}

	// 更新された行数が0の場合は失敗
	if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

// UpdateBalanceWithLock は残高を更新（悲観的ロック: SELECT FOR UPDATE）
func (ds *UserDataSourceImpl) UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

	var model UserModel

	// SELECT FOR UPDATE で行ロック
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
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
func (ds *UserDataSourceImpl) UpdateBalancesWithLock(ctx context.Context, updates []dsmysql.BalanceUpdate) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())

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
			Where("id = ?", update.UserID).
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
func (ds *UserDataSourceImpl) SelectList(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var models []UserModel

	err := db.
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
func (ds *UserDataSourceImpl) Count(ctx context.Context) (int64, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var count int64
	err := db.Model(&UserModel{}).Count(&count).Error
	return count, err
}

// Delete はユーザーを物理削除（アーカイブ後に使用）
func (ds *UserDataSourceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	return db.Delete(&UserModel{}, "id = ?", id).Error
}
