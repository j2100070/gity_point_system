package web

import "time"

// TimeProvider は時刻情報を提供するインターフェース
// 外界の一部としてみなし、このレイヤー以外では現在時刻を取得しないように制限
type TimeProvider interface {
	Now() time.Time
}

// SystemTimeProvider はシステム時刻を提供する実装
type SystemTimeProvider struct{}

// NewSystemTimeProvider は新しいSystemTimeProviderを作成
func NewSystemTimeProvider() TimeProvider {
	return &SystemTimeProvider{}
}

// Now は現在時刻を返す
func (p *SystemTimeProvider) Now() time.Time {
	return time.Now()
}

// FixedTimeProvider はテスト用の固定時刻を提供する実装
type FixedTimeProvider struct {
	fixedTime time.Time
}

// NewFixedTimeProvider は新しいFixedTimeProviderを作成
func NewFixedTimeProvider(fixedTime time.Time) TimeProvider {
	return &FixedTimeProvider{fixedTime: fixedTime}
}

// Now は固定時刻を返す
func (p *FixedTimeProvider) Now() time.Time {
	return p.fixedTime
}
