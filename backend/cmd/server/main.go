package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/config"
	"github.com/gity/point-system/internal/domain"
	"github.com/gity/point-system/internal/infrastructure/persistence"
	"github.com/gity/point-system/internal/interface/handler"
	"github.com/gity/point-system/internal/interface/middleware"
	"github.com/gity/point-system/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 設定読み込み
	cfg := config.LoadConfig()

	// データベース接続
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 依存性注入
	deps := initDependencies(db)

	// ルーター設定
	router := setupRouter(cfg, deps)

	// サーバー起動
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s (env: %s)", addr, cfg.Server.Env)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabase はデータベース接続を初期化
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	// GORM設定
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if cfg.Server.Env == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	// PostgreSQL接続
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// コネクションプール設定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")

	return db, nil
}

// Dependencies は依存性を管理
type Dependencies struct {
	// Repositories
	UserRepo         domain.UserRepository
	TransactionRepo  domain.TransactionRepository
	IdempotencyRepo  domain.IdempotencyKeyRepository
	SessionRepo      domain.SessionRepository
	FriendshipRepo   domain.FriendshipRepository
	QRCodeRepo       domain.QRCodeRepository

	// UseCases
	AuthUC          *usecase.AuthUseCase
	PointTransferUC *usecase.PointTransferUseCase
	QRCodeUC        *usecase.QRCodeUseCase
	FriendshipUC    *usecase.FriendshipUseCase
	AdminUC         *usecase.AdminUseCase

	// Handlers
	AuthHandler   *handler.AuthHandler
	PointHandler  *handler.PointHandler
	QRCodeHandler *handler.QRCodeHandler
	FriendHandler *handler.FriendHandler
	AdminHandler  *handler.AdminHandler

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware
	CSRFMiddleware *middleware.CSRFMiddleware
}

// initDependencies は依存性注入を行う
func initDependencies(db *gorm.DB) *Dependencies {
	deps := &Dependencies{}

	// Repositories
	deps.UserRepo = persistence.NewUserRepository(db)
	deps.TransactionRepo = persistence.NewTransactionRepository(db)
	deps.IdempotencyRepo = persistence.NewIdempotencyKeyRepository(db)
	deps.SessionRepo = persistence.NewSessionRepository(db)
	deps.FriendshipRepo = persistence.NewFriendshipRepository(db)
	deps.QRCodeRepo = persistence.NewQRCodeRepository(db)

	// UseCases
	deps.AuthUC = usecase.NewAuthUseCase(deps.UserRepo, deps.SessionRepo)
	deps.PointTransferUC = usecase.NewPointTransferUseCase(
		db,
		deps.UserRepo,
		deps.TransactionRepo,
		deps.IdempotencyRepo,
		deps.FriendshipRepo,
	)
	deps.QRCodeUC = usecase.NewQRCodeUseCase(deps.QRCodeRepo, deps.PointTransferUC)
	deps.FriendshipUC = usecase.NewFriendshipUseCase(deps.FriendshipRepo, deps.UserRepo)
	deps.AdminUC = usecase.NewAdminUseCase(db, deps.UserRepo, deps.TransactionRepo)

	// Handlers
	deps.AuthHandler = handler.NewAuthHandler(deps.AuthUC)
	deps.PointHandler = handler.NewPointHandler(deps.PointTransferUC)
	deps.QRCodeHandler = handler.NewQRCodeHandler(deps.QRCodeUC)
	deps.FriendHandler = handler.NewFriendHandler(deps.FriendshipUC)
	deps.AdminHandler = handler.NewAdminHandler(deps.AdminUC)

	// Middleware
	deps.AuthMiddleware = middleware.NewAuthMiddleware(deps.AuthUC)
	deps.CSRFMiddleware = middleware.NewCSRFMiddleware()

	return deps
}

// setupRouter はルーターを設定
func setupRouter(cfg *config.Config, deps *Dependencies) *gin.Engine {
	// Ginモード設定
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS設定
	corsConfig := cors.Config{
		AllowOrigins:     cfg.Security.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// セキュリティヘッダー
	router.Use(middleware.SecurityHeadersMiddleware())

	// 入力サニタイゼーション
	router.Use(middleware.InputSanitizationMiddleware())

	// ヘルスチェック
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// APIルート
	api := router.Group("/api")
	{
		// 認証（公開）
		auth := api.Group("/auth")
		{
			auth.POST("/register", deps.AuthHandler.Register)
			auth.POST("/login", deps.AuthHandler.Login)
			auth.POST("/logout", deps.AuthHandler.Logout)
			auth.GET("/me", deps.AuthHandler.GetCurrentUser)
		}

		// 認証が必要なルート
		protected := api.Group("")
		protected.Use(deps.AuthMiddleware.Authenticate())
		protected.Use(deps.CSRFMiddleware.Protect())
		{
			// ポイント
			points := protected.Group("/points")
			{
				points.POST("/transfer", deps.PointHandler.Transfer)
				points.GET("/balance", deps.PointHandler.GetBalance)
				points.GET("/history", deps.PointHandler.GetTransactionHistory)
			}

			// QRコード
			qr := protected.Group("/qr")
			{
				qr.POST("/generate/receive", deps.QRCodeHandler.GenerateReceiveQR)
				qr.POST("/generate/send", deps.QRCodeHandler.GenerateSendQR)
				qr.POST("/scan", deps.QRCodeHandler.ScanQR)
				qr.GET("/history", deps.QRCodeHandler.GetQRCodeHistory)
			}

			// 友達
			friends := protected.Group("/friends")
			{
				friends.POST("/request", deps.FriendHandler.SendFriendRequest)
				friends.POST("/accept", deps.FriendHandler.AcceptFriendRequest)
				friends.POST("/reject", deps.FriendHandler.RejectFriendRequest)
				friends.GET("", deps.FriendHandler.GetFriends)
				friends.GET("/pending", deps.FriendHandler.GetPendingRequests)
				friends.DELETE("/remove", deps.FriendHandler.RemoveFriend)
			}

			// 管理者（管理者権限必要）
			admin := protected.Group("/admin")
			admin.Use(deps.AuthMiddleware.RequireAdmin())
			{
				admin.POST("/points/grant", deps.AdminHandler.GrantPoints)
				admin.POST("/points/deduct", deps.AdminHandler.DeductPoints)
				admin.GET("/users", deps.AdminHandler.ListAllUsers)
				admin.GET("/transactions", deps.AdminHandler.ListAllTransactions)
				admin.POST("/users/role", deps.AdminHandler.UpdateUserRole)
				admin.POST("/users/deactivate", deps.AdminHandler.DeactivateUser)
			}
		}
	}

	return router
}
