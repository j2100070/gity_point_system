//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProductExchangeInteractor_ExchangeProduct は商品交換の統合テスト
func TestProductExchangeInteractor_ExchangeProduct(t *testing.T) {
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := inframysql.NewGormTransactionManager(db.GetDB())

	productExchangeUC := interactor.NewProductExchangeInteractor(
		txManager, repos.Product, repos.ProductExchange, repos.User, repos.Transaction, repos.PointBatch, lg,
	)

	// テストデータ準備
	randomID := uuid.New()
	testUser := &entities.User{
		ID:             randomID,
		Username:       "test_exchange_user_" + randomID.String()[:8],
		Email:          "exchange_" + randomID.String()[:8] + "@example.com",
		PasswordHash:   "$2a$10$test",
		DisplayName:    "Exchange Test User",
		FirstName:      "Test",
		LastName:       "User",
		Balance:        5000,
		Role:           entities.RoleUser,
		IsActive:       true,
		PersonalQRCode: "qr_" + randomID.String()[:8],
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err := repos.User.Create(context.Background(), testUser)
	require.NoError(t, err)

	// 商品作成
	testProduct, err := entities.NewProduct(
		"統合テスト商品",
		"統合テスト用の商品",
		"snack",
		200,
		10,
	)
	require.NoError(t, err)
	err = repos.Product.Create(context.Background(), testProduct)
	require.NoError(t, err)

	t.Run("正常な商品交換", func(t *testing.T) {
		resp, err := productExchangeUC.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID:    testUser.ID,
			ProductID: testProduct.ID,
			Quantity:  2,
			Notes:     "統合テスト",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, entities.ExchangeStatusCompleted, resp.Exchange.Status)
		assert.Equal(t, int64(400), resp.Exchange.PointsUsed)
		assert.Equal(t, 2, resp.Exchange.Quantity)

		// ユーザー残高確認
		updatedUser, err := repos.User.Read(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(4600), updatedUser.Balance, "残高が正しく減算されること")

		// 商品在庫確認
		updatedProduct, err := repos.Product.Read(context.Background(), testProduct.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, updatedProduct.Stock, "在庫が正しく減算されること")
	})

	t.Run("残高不足エラー", func(t *testing.T) {
		poorUserID := uuid.New()
		poorUser := &entities.User{
			ID:             poorUserID,
			Username:       "poor_user_" + poorUserID.String()[:8],
			Email:          "poor_" + poorUserID.String()[:8] + "@example.com",
			PasswordHash:   "$2a$10$test",
			DisplayName:    "Poor User",
			FirstName:      "Poor",
			LastName:       "User",
			Balance:        50,
			Role:           entities.RoleUser,
			IsActive:       true,
			PersonalQRCode: "qr_poor_" + poorUserID.String()[:8],
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err := repos.User.Create(context.Background(), poorUser)
		require.NoError(t, err)

		_, err = productExchangeUC.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID:    poorUser.ID,
			ProductID: testProduct.ID,
			Quantity:  1,
			Notes:     "残高不足テスト",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("在庫不足エラー", func(t *testing.T) {
		lowStockProduct, err := entities.NewProduct("在庫少商品", "在庫が少ない商品", "drink", 100, 1)
		require.NoError(t, err)
		err = repos.Product.Create(context.Background(), lowStockProduct)
		require.NoError(t, err)

		_, err = productExchangeUC.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID:    testUser.ID,
			ProductID: lowStockProduct.ID,
			Quantity:  5,
			Notes:     "在庫不足テスト",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})
}

// TestProductManagementInteractor は商品管理の統合テスト
func TestProductManagementInteractor(t *testing.T) {
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)

	productManagementUC := interactor.NewProductManagementInteractor(repos.Product, lg)

	t.Run("商品作成", func(t *testing.T) {
		resp, err := productManagementUC.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name:        "統合テスト新商品",
			Description: "統合テストで作成した商品",
			Category:    "snack",
			Price:       150,
			Stock:       20,
			ImageURL:    "https://example.com/test.jpg",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "統合テスト新商品", resp.Product.Name)
		assert.Equal(t, int64(150), resp.Product.Price)
		assert.True(t, resp.Product.IsAvailable)
	})

	t.Run("商品更新", func(t *testing.T) {
		createResp, err := productManagementUC.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name: "更新対象商品", Description: "更新される商品", Category: "drink", Price: 100, Stock: 10,
		})
		require.NoError(t, err)

		updateResp, err := productManagementUC.UpdateProduct(context.Background(), &inputport.UpdateProductRequest{
			ProductID: createResp.Product.ID, Name: "更新された商品", Description: "説明も更新",
			Category: "drink", Price: 120, Stock: 15, IsAvailable: false,
		})
		require.NoError(t, err)
		assert.Equal(t, "更新された商品", updateResp.Product.Name)
		assert.Equal(t, int64(120), updateResp.Product.Price)
		assert.False(t, updateResp.Product.IsAvailable)
	})

	t.Run("商品削除", func(t *testing.T) {
		createResp, err := productManagementUC.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name: "削除対象商品", Description: "削除される商品", Category: "toy", Price: 500, Stock: 5,
		})
		require.NoError(t, err)

		err = productManagementUC.DeleteProduct(context.Background(), &inputport.DeleteProductRequest{
			ProductID: createResp.Product.ID,
		})
		require.NoError(t, err)

		_, err = repos.Product.Read(context.Background(), createResp.Product.ID)
		assert.Error(t, err)
	})

	t.Run("商品一覧取得", func(t *testing.T) {
		resp, err := productManagementUC.GetProductList(context.Background(), &inputport.GetProductListRequest{
			Category: "", AvailableOnly: false, Offset: 0, Limit: 10,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("カテゴリフィルタ", func(t *testing.T) {
		_, err := productManagementUC.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name: "スナック商品", Description: "お菓子", Category: "snack", Price: 100, Stock: 10,
		})
		require.NoError(t, err)

		resp, err := productManagementUC.GetProductList(context.Background(), &inputport.GetProductListRequest{
			Category: "snack", AvailableOnly: false, Offset: 0, Limit: 10,
		})
		require.NoError(t, err)
		for _, product := range resp.Products {
			assert.Equal(t, "snack", product.CategoryCode)
		}
	})
}
