package web

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/frameworks/web/middleware"
)

// RouterConfig はルーター設定
type RouterConfig struct {
	Env            string
	AllowedOrigins []string
}

// Router はHTTPルーター
type Router struct {
	engine       *gin.Engine
	timeProvider TimeProvider
}

// NewRouter は新しいRouterを作成
func NewRouter(cfg *RouterConfig, timeProvider TimeProvider) *Router {
	// Ginモード設定
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	// CORS設定
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	engine.Use(cors.New(corsConfig))

	// セキュリティヘッダー
	engine.Use(middleware.SecurityHeadersMiddleware())

	// 入力サニタイゼーション
	engine.Use(middleware.InputSanitizationMiddleware())

	// ヘルスチェック
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Router{
		engine:       engine,
		timeProvider: timeProvider,
	}
}

// RegisterRoutes はルートを登録
// HTTP RequestのURLなどを参照し、該当するControllerへRequestを渡す
func (r *Router) RegisterRoutes(
	authController *web.AuthController,
	pointController *web.PointController,
	friendController *web.FriendController,
	qrcodeController *web.QRCodeController,
	adminController *web.AdminController,
	productController *web.ProductController,
	categoryController *web.CategoryController,
	userSettingsController *web.UserSettingsController,
	authMiddleware *middleware.AuthMiddleware,
	csrfMiddleware *middleware.CSRFMiddleware,
) {
	api := r.engine.Group("/api")
	{
		// 認証（公開）
		auth := api.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) {
				authController.Register(c, r.timeProvider.Now())
			})
			auth.POST("/login", func(c *gin.Context) {
				authController.Login(c, r.timeProvider.Now())
			})
		}

		// 商品一覧（公開）
		api.GET("/products", productController.GetProductList)

		// カテゴリ一覧（公開）
		api.GET("/categories", categoryController.GetCategoryList)

		// 認証が必要なルート（CSRF保護なし）
		protected := api.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			// 認証済みユーザー情報取得
			protected.GET("/auth/me", func(c *gin.Context) {
				authController.GetCurrentUser(c, r.timeProvider.Now())
			})

			// プロフィール取得（GET）
			protected.GET("/settings/profile", userSettingsController.GetProfile)
		}

		// 認証 + CSRF保護が必要なルート（状態変更あり）
		protectedAuth := api.Group("/auth")
		protectedAuth.Use(authMiddleware.Authenticate())
		protectedAuth.Use(csrfMiddleware.Protect())
		{
			protectedAuth.POST("/logout", func(c *gin.Context) {
				authController.Logout(c, r.timeProvider.Now())
			})
		}

		// 認証 + CSRF保護が必要なルート
		protectedWithCSRF := api.Group("")
		protectedWithCSRF.Use(authMiddleware.Authenticate())
		protectedWithCSRF.Use(csrfMiddleware.Protect())
		{
			// ポイント
			points := protectedWithCSRF.Group("/points")
			{
				// Controllerに時刻情報を渡す
				points.POST("/transfer", func(c *gin.Context) {
					pointController.Transfer(c, r.timeProvider.Now())
				})
				points.GET("/balance", func(c *gin.Context) {
					pointController.GetBalance(c, r.timeProvider.Now())
				})
				points.GET("/history", func(c *gin.Context) {
					pointController.GetTransactionHistory(c, r.timeProvider.Now())
				})
			}

			// ユーザー検索
			protectedWithCSRF.GET("/users/search", friendController.SearchUserByUsername)

			// 友達
			friends := protectedWithCSRF.Group("/friends")
			{
				friends.POST("/requests", friendController.SendFriendRequest)
				friends.POST("/requests/:id/accept", friendController.AcceptFriendRequest)
				friends.POST("/requests/:id/reject", friendController.RejectFriendRequest)
				friends.GET("", friendController.GetFriends)
				friends.GET("/requests", friendController.GetPendingRequests)
				friends.DELETE("/:id", friendController.RemoveFriend)
			}

			// QRコード
			qrcodes := protectedWithCSRF.Group("/qrcodes")
			{
				qrcodes.POST("/receive", qrcodeController.GenerateReceiveQR)
				qrcodes.POST("/send", qrcodeController.GenerateSendQR)
				qrcodes.POST("/scan", qrcodeController.ScanQR)
				qrcodes.GET("/history", qrcodeController.GetQRCodeHistory)
			}

			// 商品交換（ユーザー）
			products := protectedWithCSRF.Group("/products")
			{
				products.POST("/exchange", productController.ExchangeProduct)
				products.GET("/exchanges/history", productController.GetExchangeHistory)
				products.POST("/exchanges/:id/cancel", productController.CancelExchange)
			}

			// ユーザー設定（状態変更のみ - GETは上のprotectedグループ）
			settings := protectedWithCSRF.Group("/settings")
			{
				settings.PUT("/profile", userSettingsController.UpdateProfile)
				settings.PUT("/username", userSettingsController.UpdateUsername)
				settings.PUT("/password", userSettingsController.ChangePassword)
				settings.POST("/avatar", userSettingsController.UploadAvatar)
				settings.DELETE("/avatar", userSettingsController.DeleteAvatar)
				settings.POST("/email/verify", userSettingsController.SendEmailVerification)
				settings.POST("/email/verify/confirm", userSettingsController.VerifyEmail)
				settings.DELETE("/account", userSettingsController.ArchiveAccount)
			}

			// 管理者
			admin := protectedWithCSRF.Group("/admin")
			{
				// ポイント管理
				admin.POST("/points/grant", adminController.GrantPoints)
				admin.POST("/points/deduct", adminController.DeductPoints)

				// ユーザー管理
				admin.GET("/users", adminController.ListAllUsers)
				admin.PUT("/users/:id/role", adminController.UpdateUserRole)
				admin.POST("/users/:id/deactivate", adminController.DeactivateUser)

				// トランザクション管理
				admin.GET("/transactions", adminController.ListAllTransactions)

				// 商品管理
				admin.POST("/products", productController.CreateProduct)
				admin.PUT("/products/:id", productController.UpdateProduct)
				admin.DELETE("/products/:id", productController.DeleteProduct)

				// 商品交換管理
				admin.GET("/exchanges", productController.GetAllExchanges)
				admin.POST("/exchanges/:id/deliver", productController.MarkExchangeDelivered)

				// カテゴリ管理
				admin.POST("/categories", categoryController.CreateCategory)
				admin.PUT("/categories/:id", categoryController.UpdateCategory)
				admin.DELETE("/categories/:id", categoryController.DeleteCategory)
			}
		}
	}
}

// GetEngine はGinエンジンを取得
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run はサーバーを起動
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
