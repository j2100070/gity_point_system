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
			auth.POST("/logout", func(c *gin.Context) {
				authController.Logout(c, r.timeProvider.Now())
			})
			auth.GET("/me", func(c *gin.Context) {
				authController.GetCurrentUser(c, r.timeProvider.Now())
			})
		}

		// 認証が必要なルート
		protected := api.Group("")
		protected.Use(authMiddleware.Authenticate())
		protected.Use(csrfMiddleware.Protect())
		{
			// ポイント
			points := protected.Group("/points")
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

			// 友達
			friends := protected.Group("/friends")
			{
				friends.POST("/requests", friendController.SendFriendRequest)
				friends.POST("/requests/:id/accept", friendController.AcceptFriendRequest)
				friends.POST("/requests/:id/reject", friendController.RejectFriendRequest)
				friends.GET("", friendController.GetFriends)
				friends.GET("/requests", friendController.GetPendingRequests)
				friends.DELETE("/:id", friendController.RemoveFriend)
			}

			// QRコード
			qrcodes := protected.Group("/qrcodes")
			{
				qrcodes.POST("/receive", qrcodeController.GenerateReceiveQR)
				qrcodes.POST("/send", qrcodeController.GenerateSendQR)
				qrcodes.POST("/scan", qrcodeController.ScanQR)
				qrcodes.GET("/history", qrcodeController.GetQRCodeHistory)
			}

			// 管理者
			admin := protected.Group("/admin")
			{
				admin.POST("/points/grant", adminController.GrantPoints)
				admin.POST("/points/deduct", adminController.DeductPoints)
				admin.GET("/users", adminController.ListAllUsers)
				admin.GET("/transactions", adminController.ListAllTransactions)
				admin.PUT("/users/:id/role", adminController.UpdateUserRole)
				admin.POST("/users/:id/deactivate", adminController.DeactivateUser)
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
