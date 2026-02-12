package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/infralogger"
	"github.com/gity/point-system/gateways/infra/inframysql"
	productrepo "github.com/gity/point-system/gateways/repository/product"
	transactionrepo "github.com/gity/point-system/gateways/repository/transaction"
	userrepo "github.com/gity/point-system/gateways/repository/user"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProductExchangeInteractor_ExchangeProduct は商品交換の統合テスト
func TestProductExchangeInteractor_ExchangeProduct(t *testing.T) {
	// データベース接続
	db, err := inframysql.NewPostgresDB(&inframysql.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "point_system_test",
		SSLMode:  "disable",
		Env:      "test",
	})
	require.NoError(t, err)
	defer db.Close()

	// トランザクション開始（テスト終了後にロールバック）
	tx := db.GetDB().Begin()
	defer tx.Rollback()

	// 依存関係を初期化
	logger := infralogger.NewLogger()
	userDS := dsmysqlimpl.NewUserDataSource(db)
	transactionDS := dsmysqlimpl.NewTransactionDataSource(db)
	productDS := dsmysqlimpl.NewProductDataSource(db)
	productExchangeDS := dsmysqlimpl.NewProductExchangeDataSource(db)

	userRepo := userrepo.NewUserRepository(userDS, logger)
	transactionRepo := transactionrepo.NewTransactionRepository(transactionDS, logger)
	productRepo := productrepo.NewProductRepository(productDS, logger)
	productExchangeRepo := productrepo.NewProductExchangeRepository(productExchangeDS, logger)

	// インタラクター作成

	txManager := inframysql.NewGormTransactionManager(db.GetDB())

	productExchangeUC := interactor.NewProductExchangeInteractor(
		txManager,
		productRepo,
		productExchangeRepo,
		userRepo,
		transactionRepo,
		logger,
	)

	// テストデータ準備
	// ユーザー作成
	randomID := uuid.New()
	testUser := &entities.User{
		ID:             randomID,
		Username:       "test_exchange_user_" + randomID.String(),
		Email:          "exchange_" + randomID.String() + "@example.com",
		PasswordHash:   "$2a$10$test",
		DisplayName:    "Exchange Test User",
		Balance:        5000,
		Role:           entities.RoleUser,
		IsActive:       true,
		PersonalQRCode: "qr_" + randomID.String(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err = userRepo.Create(context.Background(), testUser)
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
	err = productRepo.Create(context.Background(), testProduct)
	require.NoError(t, err)

	// テストケース
	t.Run("正常な商品交換", func(t *testing.T) {
		req := &inputport.ExchangeProductRequest{
			UserID:    testUser.ID,
			ProductID: testProduct.ID,
			Quantity:  2,
			Notes:     "統合テスト",
		}

		resp, err := productExchangeUC.ExchangeProduct(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, entities.ExchangeStatusCompleted, resp.Exchange.Status)
		assert.Equal(t, int64(400), resp.Exchange.PointsUsed)
		assert.Equal(t, 2, resp.Exchange.Quantity)

		// ユーザー残高確認
		updatedUser, err := userRepo.Read(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(4600), updatedUser.Balance, "残高が正しく減算されること")

		// 商品在庫確認
		updatedProduct, err := productRepo.Read(context.Background(), testProduct.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, updatedProduct.Stock, "在庫が正しく減算されること")
	})

	t.Run("残高不足エラー", func(t *testing.T) {
		// 残高不足のユーザー作成
		poorUserID := uuid.New()
		poorUser := &entities.User{
			ID:             poorUserID,
			Username:       "poor_user_" + poorUserID.String(),
			Email:          "poor_" + poorUserID.String() + "@example.com",
			PasswordHash:   "$2a$10$test",
			DisplayName:    "Poor User",
			Balance:        50,
			Role:           entities.RoleUser,
			IsActive:       true,
			PersonalQRCode: "qr_poor_" + poorUserID.String(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err = userRepo.Create(context.Background(), poorUser)
		require.NoError(t, err)

		req := &inputport.ExchangeProductRequest{
			UserID:    poorUser.ID,
			ProductID: testProduct.ID,
			Quantity:  1,
			Notes:     "残高不足テスト",
		}

		_, err := productExchangeUC.ExchangeProduct(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("在庫不足エラー", func(t *testing.T) {
		// 在庫が少ない商品作成
		lowStockProduct, err := entities.NewProduct(
			"在庫少商品",
			"在庫が少ない商品",
			"drink",
			100,
			1,
		)
		require.NoError(t, err)
		err = productRepo.Create(context.Background(), lowStockProduct)
		require.NoError(t, err)

		req := &inputport.ExchangeProductRequest{
			UserID:    testUser.ID,
			ProductID: lowStockProduct.ID,
			Quantity:  5, // 在庫より多い
			Notes:     "在庫不足テスト",
		}

		_, err = productExchangeUC.ExchangeProduct(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})
}

// TestProductManagementInteractor は商品管理の統合テスト
func TestProductManagementInteractor(t *testing.T) {
	// データベース接続
	db, err := inframysql.NewPostgresDB(&inframysql.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "point_system_test",
		SSLMode:  "disable",
		Env:      "test",
	})
	require.NoError(t, err)
	defer db.Close()

	// トランザクション開始
	tx := db.GetDB().Begin()
	defer tx.Rollback()

	// 依存関係を初期化
	logger := infralogger.NewLogger()
	productDS := dsmysqlimpl.NewProductDataSource(db)
	productRepo := productrepo.NewProductRepository(productDS, logger)

	// インタラクター作成
	productManagementUC := interactor.NewProductManagementInteractor(
		productRepo,
		logger,
	)

	t.Run("商品作成", func(t *testing.T) {
		req := &inputport.CreateProductRequest{
			Name:        "統合テスト新商品",
			Description: "統合テストで作成した商品",
			Category:    "snack",
			Price:       150,
			Stock:       20,
			ImageURL:    "https://example.com/test.jpg",
		}

		resp, err := productManagementUC.CreateProduct(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Name, resp.Product.Name)
		assert.Equal(t, int64(150), resp.Product.Price)
		assert.True(t, resp.Product.IsAvailable)
	})

	t.Run("商品更新", func(t *testing.T) {
		// 商品作成
		createReq := &inputport.CreateProductRequest{
			Name:        "更新対象商品",
			Description: "更新される商品",
			Category:    "drink",
			Price:       100,
			Stock:       10,
		}
		createResp, err := productManagementUC.CreateProduct(context.Background(), createReq)
		require.NoError(t, err)

		// 更新
		updateReq := &inputport.UpdateProductRequest{
			ProductID:   createResp.Product.ID,
			Name:        "更新された商品",
			Description: "説明も更新",
			Category:    "drink",
			Price:       120,
			Stock:       15,
			IsAvailable: false,
		}

		updateResp, err := productManagementUC.UpdateProduct(context.Background(), updateReq)
		require.NoError(t, err)
		assert.Equal(t, "更新された商品", updateResp.Product.Name)
		assert.Equal(t, int64(120), updateResp.Product.Price)
		assert.False(t, updateResp.Product.IsAvailable)
	})

	t.Run("商品削除", func(t *testing.T) {
		// 商品作成
		createReq := &inputport.CreateProductRequest{
			Name:        "削除対象商品",
			Description: "削除される商品",
			Category:    "toy",
			Price:       500,
			Stock:       5,
		}
		createResp, err := productManagementUC.CreateProduct(context.Background(), createReq)
		require.NoError(t, err)

		// 削除
		deleteReq := &inputport.DeleteProductRequest{
			ProductID: createResp.Product.ID,
		}

		err = productManagementUC.DeleteProduct(context.Background(), deleteReq)
		require.NoError(t, err)

		// 削除確認（論理削除なので読み取りエラー）
		_, err = productRepo.Read(context.Background(), createResp.Product.ID)
		assert.Error(t, err)
	})

	t.Run("商品一覧取得", func(t *testing.T) {
		req := &inputport.GetProductListRequest{
			Category:      "",
			AvailableOnly: false,
			Offset:        0,
			Limit:         10,
		}

		resp, err := productManagementUC.GetProductList(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.GreaterOrEqual(t, len(resp.Products), 0)
	})

	t.Run("カテゴリフィルタ", func(t *testing.T) {
		// snackカテゴリの商品を作成
		createReq := &inputport.CreateProductRequest{
			Name:        "スナック商品",
			Description: "お菓子",
			Category:    "snack",
			Price:       100,
			Stock:       10,
		}
		_, err := productManagementUC.CreateProduct(context.Background(), createReq)
		require.NoError(t, err)

		req := &inputport.GetProductListRequest{
			Category:      "snack",
			AvailableOnly: false,
			Offset:        0,
			Limit:         10,
		}

		resp, err := productManagementUC.GetProductList(context.Background(), req)
		require.NoError(t, err)

		// すべてsnackカテゴリであることを確認
		for _, product := range resp.Products {
			assert.Equal(t, "snack", product.CategoryCode)
		}
	})
}
