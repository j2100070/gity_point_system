package dsmysqlimpl

import (
	"context"
	"time"

	"github.com/gity/point-system/gateways/infra/inframysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SystemSettingModel はシステム設定のGORMモデル
type SystemSettingModel struct {
	Key         string    `gorm:"type:varchar(100);primary_key"`
	Value       string    `gorm:"type:text;not null"`
	Description *string   `gorm:"type:text"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName はテーブル名を指定
func (SystemSettingModel) TableName() string {
	return "system_settings"
}

// SystemSettingsDataSource はシステム設定のデータソース
type SystemSettingsDataSource struct {
	db inframysql.DB
}

// NewSystemSettingsDataSource は新しいSystemSettingsDataSourceを作成
func NewSystemSettingsDataSource(db inframysql.DB) *SystemSettingsDataSource {
	return &SystemSettingsDataSource{db: db}
}

// GetSetting はキーに対応する設定値を取得
func (ds *SystemSettingsDataSource) GetSetting(ctx context.Context, key string) (string, error) {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	var model SystemSettingModel
	err := db.Where("key = ?", key).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return model.Value, nil
}

// SetSetting はキーに対応する設定値を保存（upsert）
func (ds *SystemSettingsDataSource) SetSetting(ctx context.Context, key, value, description string) error {
	db := inframysql.GetDB(ctx, ds.db.GetDB())
	model := SystemSettingModel{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
	if description != "" {
		model.Description = &description
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "description", "updated_at"}),
	}).Create(&model).Error
}
