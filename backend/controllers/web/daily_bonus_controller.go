package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// DailyBonusController はデイリーボーナスのコントローラー
type DailyBonusController struct {
	dailyBonusPort inputport.DailyBonusInputPort
	presenter      *presenter.DailyBonusPresenter
}

// NewDailyBonusController は新しいDailyBonusControllerを作成
func NewDailyBonusController(
	dailyBonusPort inputport.DailyBonusInputPort,
	presenter *presenter.DailyBonusPresenter,
) *DailyBonusController {
	return &DailyBonusController{
		dailyBonusPort: dailyBonusPort,
		presenter:      presenter,
	}
}

// GetTodayBonus は本日のボーナス状況を取得
func (c *DailyBonusController) GetTodayBonus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := c.dailyBonusPort.GetTodayBonus(ctx, &inputport.GetTodayBonusRequest{
		UserID: userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, c.presenter.PresentGetTodayBonus(resp))
}

// GetRecentBonuses は最近のボーナス履歴を取得
func (c *DailyBonusController) GetRecentBonuses(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 7 // デフォルトで7日分
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	resp, err := c.dailyBonusPort.GetRecentBonuses(ctx, &inputport.GetRecentBonusesRequest{
		UserID: userID.(uuid.UUID),
		Limit:  limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, c.presenter.PresentGetRecentBonuses(resp))
}

// GetBonusSettings はボーナス設定を取得（管理者用、抽選ティア含む）
func (c *DailyBonusController) GetBonusSettings(ctx *gin.Context) {
	resp, err := c.dailyBonusPort.GetBonusSettings(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tiers := make([]gin.H, len(resp.LotteryTiers))
	for i, tier := range resp.LotteryTiers {
		tiers[i] = gin.H{
			"id":            tier.ID,
			"name":          tier.Name,
			"points":        tier.Points,
			"probability":   tier.Probability,
			"display_order": tier.DisplayOrder,
			"is_active":     tier.IsActive,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bonus_points":  resp.BonusPoints,
		"lottery_tiers": tiers,
	})
}

// UpdateLotteryTiers は抽選ティアを一括更新（管理者用）
func (c *DailyBonusController) UpdateLotteryTiers(ctx *gin.Context) {
	var req struct {
		Tiers []struct {
			Name         string  `json:"name" binding:"required"`
			Points       int64   `json:"points" binding:"min=0"`
			Probability  float64 `json:"probability" binding:"required,min=0,max=100"`
			DisplayOrder int     `json:"display_order"`
		} `json:"tiers" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	tiers := make([]inputport.LotteryTierInput, len(req.Tiers))
	for i, t := range req.Tiers {
		tiers[i] = inputport.LotteryTierInput{
			Name:         t.Name,
			Points:       t.Points,
			Probability:  t.Probability,
			DisplayOrder: t.DisplayOrder,
		}
	}

	err := c.dailyBonusPort.UpdateLotteryTiers(ctx, &inputport.UpdateLotteryTiersRequest{
		Tiers: tiers,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "抽選ティア設定を更新しました",
	})
}

// MarkBonusViewed はボーナスを閲覧済みにする
func (c *DailyBonusController) MarkBonusViewed(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		BonusID string `json:"bonus_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	bonusUUID, err := uuid.Parse(req.BonusID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid bonus_id"})
		return
	}

	err = c.dailyBonusPort.MarkBonusViewed(ctx, &inputport.MarkBonusViewedRequest{
		BonusID: bonusUUID,
		UserID:  userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "閲覧済みにしました",
	})
}
