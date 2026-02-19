package service

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
)

// AkerunAccessGateway はAkerun入退室APIとの通信インターフェース
// インフラ層のAkerunClientがこのインターフェースを実装する
type AkerunAccessGateway interface {
	// FetchAccesses は指定期間のアクセス記録を取得する
	FetchAccesses(ctx context.Context, after, before time.Time, limit int) ([]entities.AccessRecord, error)
	// IsConfigured はAkerun APIが設定済みかを返す
	IsConfigured() bool
}
