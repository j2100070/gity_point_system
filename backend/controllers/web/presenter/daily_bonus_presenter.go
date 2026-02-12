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
	return map[string]interface{}{
		"daily_bonus": map[string]interface{}{
			"id":                   resp.DailyBonus.ID,
			"user_id":              resp.DailyBonus.UserID,
			"bonus_date":           resp.DailyBonus.BonusDate.Format("2006-01-02"),
			"login_completed":      resp.DailyBonus.LoginCompleted,
			"login_completed_at":   resp.DailyBonus.LoginCompletedAt,
			"transfer_completed":   resp.DailyBonus.TransferCompleted,
			"transfer_completed_at": resp.DailyBonus.TransferCompletedAt,
			"exchange_completed":   resp.DailyBonus.ExchangeCompleted,
			"exchange_completed_at": resp.DailyBonus.ExchangeCompletedAt,
			"all_completed":        resp.DailyBonus.AllCompleted,
			"all_completed_at":     resp.DailyBonus.AllCompletedAt,
			"total_bonus_points":   resp.DailyBonus.TotalBonusPoints,
			"completed_count":      resp.DailyBonus.GetCompletedCount(),
			"remaining_bonus":      resp.DailyBonus.GetRemainingBonus(),
		},
		"all_completed_count":     resp.AllCompletedCount,
		"can_claim_login_bonus":   resp.CanClaimLoginBonus,
		"can_claim_transfer_bonus": resp.CanClaimTransferBonus,
		"can_claim_exchange_bonus": resp.CanClaimExchangeBonus,
	}
}

// PresentCheckBonus はボーナスチェックレスポンスを生成（共通処理）
func (p *DailyBonusPresenter) PresentCheckBonus(dailyBonus interface{}, bonusAwarded int64, user interface{}) map[string]interface{} {
	return map[string]interface{}{
		"bonus_awarded": bonusAwarded,
		"new_balance":   user,
		"daily_bonus":   dailyBonus,
		"message":       p.getBonusMessage(bonusAwarded, dailyBonus),
	}
}

// getBonusMessage はボーナスメッセージを生成
func (p *DailyBonusPresenter) getBonusMessage(bonusAwarded int64, dailyBonus interface{}) string {
	if bonusAwarded == 0 {
		return "本日のこのボーナスは既に獲得済みです"
	}
	if bonusAwarded == 20 {
		// 全達成ボーナス（追加分）が付与された
		return "🎉 本日のデイリーボーナスを全て達成しました！+50Pゲット！"
	}
	if bonusAwarded == 10 {
		return "+10P ボーナスゲット！"
	}
	return "ボーナスを獲得しました！"
}

// PresentGetRecentBonuses は最近のボーナス履歴レスポンスを生成
func (p *DailyBonusPresenter) PresentGetRecentBonuses(resp *inputport.GetRecentBonusesResponse) map[string]interface{} {
	bonuses := make([]map[string]interface{}, len(resp.Bonuses))
	for i, bonus := range resp.Bonuses {
		bonuses[i] = map[string]interface{}{
			"id":                   bonus.ID,
			"bonus_date":           bonus.BonusDate.Format("2006-01-02"),
			"login_completed":      bonus.LoginCompleted,
			"transfer_completed":   bonus.TransferCompleted,
			"exchange_completed":   bonus.ExchangeCompleted,
			"all_completed":        bonus.AllCompleted,
			"total_bonus_points":   bonus.TotalBonusPoints,
			"completed_count":      bonus.GetCompletedCount(),
		}
	}

	return map[string]interface{}{
		"bonuses":             bonuses,
		"all_completed_count": resp.AllCompletedCount,
	}
}
