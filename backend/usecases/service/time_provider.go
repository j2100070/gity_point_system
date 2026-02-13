package service

import "time"

// TimeProvider は時刻情報を提供するサービスインターフェース
// テスタビリティのために現在時刻を抽象化
type TimeProvider interface {
	// Now は現在時刻を返す
	Now() time.Time
}
