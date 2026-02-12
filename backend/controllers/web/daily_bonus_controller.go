package web

import (
	"net/http"
	"strconv"
	"time"

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

	resp, err := c.dailyBonusPort.GetTodayBonus(&inputport.GetTodayBonusRequest{
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

	resp, err := c.dailyBonusPort.GetRecentBonuses(&inputport.GetRecentBonusesRequest{
		UserID: userID.(uuid.UUID),
		Limit:  limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, c.presenter.PresentGetRecentBonuses(resp))
}

// ClaimLoginBonus はログインボーナスを請求
func (c *DailyBonusController) ClaimLoginBonus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := c.dailyBonusPort.CheckLoginBonus(&inputport.CheckLoginBonusRequest{
		UserID: userID.(uuid.UUID),
		Date:   time.Now(),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, c.presenter.PresentCheckBonus(
		resp.DailyBonus,
		resp.BonusAwarded,
		map[string]interface{}{
			"id":       resp.User.ID,
			"username": resp.User.Username,
			"balance":  resp.User.Balance,
		},
	))
}

