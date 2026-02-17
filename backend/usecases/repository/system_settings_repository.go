package repository

import "context"

// SystemSettingsRepository はシステム設定のリポジトリインターフェース
type SystemSettingsRepository interface {
	// GetSetting はキーに対応する設定値を取得
	GetSetting(ctx context.Context, key string) (string, error)

	// SetSetting はキーに対応する設定値を保存
	SetSetting(ctx context.Context, key, value, description string) error
}
