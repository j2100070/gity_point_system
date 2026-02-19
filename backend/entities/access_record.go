package entities

import (
	"time"

	"github.com/google/uuid"
)

// AccessRecord はAkerun入退室記録のドメインDTO
// インフラ層のAkerun API固有の構造体から変換されて渡される
type AccessRecord struct {
	ID         uuid.UUID // Akerunアクセス記録ID
	UserName   string    // Akerunユーザー名
	AccessedAt time.Time // アクセス時刻（パース済み）
}
