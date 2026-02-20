package presenter

import (
	"github.com/gity/point-system/usecases/inputport"
)

// DailyBonusPresenter はデイリーボーナスのプレゼンター
type DailyBonusPresenter struct{}

// NewDailyBonusPresenter は新しいDailyBonusPresenterを作成
func NewDailyBonusPresenter() *DailyBonusPresenter {
	return &DailyBonusPresenter{}
}

// PresentGetTodayBonus は本日のボーナス状況レスポンスを生成
func (p *DailyBonusPresenter) PresentGetTodayBonus(resp *inputport.GetTodayBonusResponse) map[string]interface{} {
	result := map[string]interface{}{
		"claimed":            resp.DailyBonus != nil,
		"bonus_points":       resp.BonusPoints,
		"total_days":         resp.TotalDays,
		"is_lottery_pending": resp.IsLotteryPending,
	}

	if resp.DailyBonus != nil {
		result["daily_bonus"] = map[string]interface{}{
			"id":                resp.DailyBonus.ID,
			"user_id":           resp.DailyBonus.UserID,
			"bonus_date":        resp.DailyBonus.BonusDate.Format("2006-01-02"),
			"bonus_points":      resp.DailyBonus.BonusPoints,
			"akerun_user_name":  resp.DailyBonus.AkerunUserName,
			"accessed_at":       resp.DailyBonus.AccessedAt,
			"lottery_tier_name": resp.DailyBonus.LotteryTierName,
			"is_viewed":         resp.DailyBonus.IsViewed,
			"is_drawn":          resp.DailyBonus.IsDrawn,
			"created_at":        resp.DailyBonus.CreatedAt,
		}
	}

	return result
}

// PresentGetRecentBonuses は最近のボーナス履歴レスポンスを生成
func (p *DailyBonusPresenter) PresentGetRecentBonuses(resp *inputport.GetRecentBonusesResponse) map[string]interface{} {
	bonuses := make([]map[string]interface{}, len(resp.Bonuses))
	for i, bonus := range resp.Bonuses {
		bonuses[i] = map[string]interface{}{
			"id":                bonus.ID,
			"bonus_date":        bonus.BonusDate.Format("2006-01-02"),
			"bonus_points":      bonus.BonusPoints,
			"akerun_user_name":  bonus.AkerunUserName,
			"accessed_at":       bonus.AccessedAt,
			"lottery_tier_name": bonus.LotteryTierName,
			"is_drawn":          bonus.IsDrawn,
		}
	}

	return map[string]interface{}{
		"bonuses":    bonuses,
		"total_days": resp.TotalDays,
	}
}
