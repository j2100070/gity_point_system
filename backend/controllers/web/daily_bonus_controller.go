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

// GetBonusSettings はボーナス設定を取得（管理者用）
func (c *DailyBonusController) GetBonusSettings(ctx *gin.Context) {
	resp, err := c.dailyBonusPort.GetBonusSettings(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bonus_points": resp.BonusPoints,
	})
}

// UpdateBonusSettings はボーナス設定を更新（管理者用）
func (c *DailyBonusController) UpdateBonusSettings(ctx *gin.Context) {
	var req struct {
		BonusPoints int64 `json:"bonus_points" binding:"required,min=1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: bonus_points must be >= 1"})
		return
	}

	err := c.dailyBonusPort.UpdateBonusSettings(ctx, &inputport.UpdateBonusSettingsRequest{
		BonusPoints: req.BonusPoints,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "ボーナスポイント設定を更新しました",
		"bonus_points": req.BonusPoints,
	})
}
