package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// ProductController は商品管理のコントローラー
type ProductController struct {
	productManagementUseCase inputport.ProductManagementInputPort
	productExchangeUseCase   inputport.ProductExchangeInputPort
	logger                   entities.Logger
}

// NewProductController は新しいProductControllerを作成
func NewProductController(
	productManagementUseCase inputport.ProductManagementInputPort,
	productExchangeUseCase inputport.ProductExchangeInputPort,
	logger entities.Logger,
) *ProductController {
	return &ProductController{
		productManagementUseCase: productManagementUseCase,
		productExchangeUseCase:   productExchangeUseCase,
		logger:                   logger,
	}
}

// CreateProduct は新しい商品を作成（管理者のみ）
// POST /admin/products
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	var req inputport.CreateProductRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.productManagementUseCase.CreateProduct(&req)
	if err != nil {
		c.logger.Error("Failed to create product", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// UpdateProduct は商品情報を更新（管理者のみ）
// PUT /admin/products/:id
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	productID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req inputport.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ProductID = productID

	resp, err := c.productManagementUseCase.UpdateProduct(&req)
	if err != nil {
		c.logger.Error("Failed to update product", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteProduct は商品を削除（管理者のみ）
// DELETE /admin/products/:id
func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	productID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	req := &inputport.DeleteProductRequest{
		ProductID: productID,
	}

	if err := c.productManagementUseCase.DeleteProduct(req); err != nil {
		c.logger.Error("Failed to delete product", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

// GetProductList は商品一覧を取得
// GET /products?category=snack&available_only=true&offset=0&limit=20
func (c *ProductController) GetProductList(ctx *gin.Context) {
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	category := ctx.Query("category")
	availableOnly := ctx.Query("available_only") == "true"

	req := &inputport.GetProductListRequest{
		Category:      category,
		AvailableOnly: availableOnly,
		Offset:        offset,
		Limit:         limit,
	}

	resp, err := c.productManagementUseCase.GetProductList(req)
	if err != nil {
		c.logger.Error("Failed to get product list", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExchangeProduct はポイントで商品を交換
// POST /products/exchange
func (c *ProductController) ExchangeProduct(ctx *gin.Context) {
	// ユーザーIDはセッションから取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var reqBody struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
		Notes     string `json:"notes"`
	}

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productID, err := uuid.Parse(reqBody.ProductID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	req := &inputport.ExchangeProductRequest{
		UserID:    userID.(uuid.UUID),
		ProductID: productID,
		Quantity:  reqBody.Quantity,
		Notes:     reqBody.Notes,
	}

	resp, err := c.productExchangeUseCase.ExchangeProduct(req)
	if err != nil {
		c.logger.Error("Failed to exchange product", entities.NewField("error", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetExchangeHistory は交換履歴を取得
// GET /products/exchanges/history?offset=0&limit=20
func (c *ProductController) GetExchangeHistory(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	req := &inputport.GetExchangeHistoryRequest{
		UserID: userID.(uuid.UUID),
		Offset: offset,
		Limit:  limit,
	}

	resp, err := c.productExchangeUseCase.GetExchangeHistory(req)
	if err != nil {
		c.logger.Error("Failed to get exchange history", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CancelExchange は交換をキャンセル
// POST /products/exchanges/:id/cancel
func (c *ProductController) CancelExchange(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	exchangeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange ID"})
		return
	}

	req := &inputport.CancelExchangeRequest{
		UserID:     userID.(uuid.UUID),
		ExchangeID: exchangeID,
	}

	if err := c.productExchangeUseCase.CancelExchange(req); err != nil {
		c.logger.Error("Failed to cancel exchange", entities.NewField("error", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "exchange cancelled successfully"})
}

// MarkExchangeDelivered は配達完了にする（管理者のみ）
// POST /admin/exchanges/:id/deliver
func (c *ProductController) MarkExchangeDelivered(ctx *gin.Context) {
	exchangeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange ID"})
		return
	}

	req := &inputport.MarkExchangeDeliveredRequest{
		ExchangeID: exchangeID,
	}

	if err := c.productExchangeUseCase.MarkExchangeDelivered(req); err != nil {
		c.logger.Error("Failed to mark exchange as delivered", entities.NewField("error", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "exchange marked as delivered"})
}

// GetAllExchanges はすべての交換履歴を取得（管理者のみ）
// GET /admin/exchanges?offset=0&limit=20
func (c *ProductController) GetAllExchanges(ctx *gin.Context) {
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	resp, err := c.productExchangeUseCase.GetAllExchanges(offset, limit)
	if err != nil {
		c.logger.Error("Failed to get all exchanges", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
