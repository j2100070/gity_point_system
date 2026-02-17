package system_settings

import (
	"context"

	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
)

// SystemSettingsRepositoryImpl はシステム設定リポジトリの実装
type SystemSettingsRepositoryImpl struct {
	ds *dsmysqlimpl.SystemSettingsDataSource
}

// NewSystemSettingsRepository は新しいSystemSettingsRepositoryを作成
func NewSystemSettingsRepository(ds *dsmysqlimpl.SystemSettingsDataSource) *SystemSettingsRepositoryImpl {
	return &SystemSettingsRepositoryImpl{ds: ds}
}

// GetSetting はキーに対応する設定値を取得
func (r *SystemSettingsRepositoryImpl) GetSetting(ctx context.Context, key string) (string, error) {
	return r.ds.GetSetting(ctx, key)
}

// SetSetting はキーに対応する設定値を保存
func (r *SystemSettingsRepositoryImpl) SetSetting(ctx context.Context, key, value, description string) error {
	return r.ds.SetSetting(ctx, key, value, description)
}
