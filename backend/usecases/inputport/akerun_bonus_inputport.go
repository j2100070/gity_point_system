package inputport

import (
	"context"
	"time"

	"github.com/gity/point-system/entities"
)

// AkerunBonusInputPort はAkerunアクセス記録からのボーナス付与ユースケースインターフェース
// AkerunWorker（インフラ層）からビジネスロジックを分離して呼び出すために使用する
type AkerunBonusInputPort interface {
	// ProcessAccesses はアクセス記録を処理してボーナスを付与する
	ProcessAccesses(ctx context.Context, accesses []entities.AccessRecord) error
	// GetLastPolledAt は前回ポーリング時刻を取得する
	GetLastPolledAt(ctx context.Context) (time.Time, error)
	// UpdateLastPolledAt はポーリング時刻を更新する
	UpdateLastPolledAt(ctx context.Context, t time.Time) error
}
