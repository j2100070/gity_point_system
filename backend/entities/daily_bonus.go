package entities

import (
	"time"

	"github.com/google/uuid"
)

// DefaultAkerunBonusPoints はAkerun入退室ボーナスのデフォルトポイント数
const DefaultAkerunBonusPoints int64 = 5

// DailyBonus はAkerun入退室ベースのデイリーボーナスエンティティ
type DailyBonus struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	BonusDate      time.Time // JST AM6:00区切りの日付
	BonusPoints    int64
	AkerunAccessID string
	AkerunUserName string
	AccessedAt     *time.Time
	CreatedAt      time.Time
}

// NewDailyBonus は新しいDailyBonusを作成
func NewDailyBonus(userID uuid.UUID, bonusDate time.Time, bonusPoints int64, akerunAccessID, akerunUserName string, accessedAt *time.Time) *DailyBonus {
	return &DailyBonus{
		ID:             uuid.New(),
		UserID:         userID,
		BonusDate:      bonusDate,
		BonusPoints:    bonusPoints,
		AkerunAccessID: akerunAccessID,
		AkerunUserName: akerunUserName,
		AccessedAt:     accessedAt,
		CreatedAt:      time.Now(),
	}
}

// GetBonusDateJST はJST AM6:00区切りでボーナス対象日を計算する
// AM6:00より前の場合は前日扱い
func GetBonusDateJST(t time.Time) time.Time {
	jst := time.FixedZone("JST", 9*60*60)
	tJST := t.In(jst)

	// AM6:00より前なら前日扱い
	if tJST.Hour() < 6 {
		tJST = tJST.AddDate(0, 0, -1)
	}

	return time.Date(tJST.Year(), tJST.Month(), tJST.Day(), 0, 0, 0, 0, jst)
}
