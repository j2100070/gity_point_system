package infratime

import (
	"time"

	"github.com/gity/point-system/usecases/service"
)

// SystemTimeProvider はシステム時刻を提供する実装
type SystemTimeProvider struct{}

// NewSystemTimeProvider は新しいSystemTimeProviderを作成
func NewSystemTimeProvider() service.TimeProvider {
	return &SystemTimeProvider{}
}

// Now は現在時刻を返す
func (p *SystemTimeProvider) Now() time.Time {
	return time.Now()
}
