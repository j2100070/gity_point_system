package main

import (
	"fmt"
	"log"

	"github.com/gity/point-system/config"
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/infraemail"
	"github.com/gity/point-system/gateways/infra/infralogger"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/infra/infrapassword"
	"github.com/gity/point-system/gateways/infra/infrastorage"
	categoryrepo "github.com/gity/point-system/gateways/repository/category"
	friendshiprepo "github.com/gity/point-system/gateways/repository/friendship"
	productrepo "github.com/gity/point-system/gateways/repository/product"
	qrcoderepo "github.com/gity/point-system/gateways/repository/qrcode"
	sessionrepo "github.com/gity/point-system/gateways/repository/session"
	transactionrepo "github.com/gity/point-system/gateways/repository/transaction"
	userrepo "github.com/gity/point-system/gateways/repository/user"
	usersettingsrepo "github.com/gity/point-system/gateways/repository/user_settings"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/frameworks/web/middleware"
	"github.com/gity/point-system/usecases/interactor"
)

// AppContainer はアプリケーションの依存関係を管理
type AppContainer struct {
	Router *frameworksweb.Router
	DB     inframysql.DB
}

// NewAppContainer は新しいAppContainerを作成（手動DI）
func NewAppContainer(dbConfig *inframysql.Config, routerConfig *frameworksweb.RouterConfig) (*AppContainer, error) {
	// === Infra層 ===
	db, err := inframysql.NewPostgresDB(dbConfig)
	if err != nil {
		return nil, err
	}

	logger := infralogger.NewLogger()
	logger.Info("Database connection established")

	// === AutoMigrate（新規テーブルのみ自動作成） ===
	// 既存テーブルはSQLマイグレーション(migrations/*.sql)で管理
	// 新規追加したモデルのみここに記載する
	if err := db.GetDB().AutoMigrate(
		&dsmysqlimpl.CategoryModel{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}
	logger.Info("Database auto migration completed")

	// === DataSource層 ===
	userDS := dsmysqlimpl.NewUserDataSource(db)
	transactionDS := dsmysqlimpl.NewTransactionDataSource(db)
	idempotencyDS := dsmysqlimpl.NewIdempotencyKeyDataSource(db)
	sessionDS := dsmysqlimpl.NewSessionDataSource(db)
	friendshipDS := dsmysqlimpl.NewFriendshipDataSource(db)
	qrcodeDS := dsmysqlimpl.NewQRCodeDataSource(db)
	productDS := dsmysqlimpl.NewProductDataSource(db)
	productExchangeDS := dsmysqlimpl.NewProductExchangeDataSource(db)
	categoryDS := dsmysqlimpl.NewCategoryDataSource(db)
	archivedUserDS := dsmysqlimpl.NewArchivedUserDataSource(db)
	emailVerificationDS := dsmysqlimpl.NewEmailVerificationDataSource(db)
	usernameChangeHistoryDS := dsmysqlimpl.NewUsernameChangeHistoryDataSource(db)
	passwordChangeHistoryDS := dsmysqlimpl.NewPasswordChangeHistoryDataSource(db)

	// === Repository層 ===
	userRepo := userrepo.NewUserRepository(userDS, logger)
	transactionRepo := transactionrepo.NewTransactionRepository(transactionDS, logger)
	idempotencyRepo := transactionrepo.NewIdempotencyKeyRepository(idempotencyDS, logger)
	sessionRepo := sessionrepo.NewSessionRepository(sessionDS, logger)
	friendshipRepo := friendshiprepo.NewFriendshipRepository(friendshipDS, logger)
	qrcodeRepo := qrcoderepo.NewQRCodeRepository(qrcodeDS, logger)
	productRepo := productrepo.NewProductRepository(productDS, logger)
	productExchangeRepo := productrepo.NewProductExchangeRepository(productExchangeDS, logger)
	categoryRepo := categoryrepo.NewCategoryRepository(categoryDS, logger)
	userSettingsRepo := usersettingsrepo.NewUserSettingsRepository(userDS, logger)
	archivedUserRepo := usersettingsrepo.NewArchivedUserRepository(archivedUserDS, logger)
	emailVerificationRepo := usersettingsrepo.NewEmailVerificationRepository(emailVerificationDS, logger)
	usernameChangeHistoryRepo := usersettingsrepo.NewUsernameChangeHistoryRepository(usernameChangeHistoryDS, logger)
	passwordChangeHistoryRepo := usersettingsrepo.NewPasswordChangeHistoryRepository(passwordChangeHistoryDS, logger)

	// === Interactor層 ===
	authUC := interactor.NewAuthInteractor(
		userRepo,
		sessionRepo,
		logger,
	)

	pointTransferUC := interactor.NewPointTransferInteractor(
		db.GetDB(),
		userRepo,
		transactionRepo,
		idempotencyRepo,
		friendshipRepo,
		logger,
	)

	friendshipUC := interactor.NewFriendshipInteractor(
		friendshipRepo,
		userRepo,
		logger,
	)

	qrcodeUC := interactor.NewQRCodeInteractor(
		qrcodeRepo,
		pointTransferUC,
		logger,
	)

	adminUC := interactor.NewAdminInteractor(
		db.GetDB(),
		userRepo,
		transactionRepo,
		idempotencyRepo,
		logger,
	)

	productManagementUC := interactor.NewProductManagementInteractor(
		productRepo,
		logger,
	)

	productExchangeUC := interactor.NewProductExchangeInteractor(
		db.GetDB(),
		productRepo,
		productExchangeRepo,
		userRepo,
		transactionRepo,
		logger,
	)

	categoryUC := interactor.NewCategoryManagementInteractor(
		categoryRepo,
		logger,
	)

	// === Service層 ===
	fileStorageService, err := infrastorage.NewLocalStorage(&infrastorage.Config{
		BaseDir:   "./uploads/avatars",
		BaseURL:   "/uploads/avatars",
		MaxSizeMB: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file storage service: %w", err)
	}
	passwordService := infrapassword.NewBcryptPasswordService()
	emailService := infraemail.NewConsoleEmailService(logger)

	userSettingsUC := interactor.NewUserSettingsInteractor(
		db.GetDB(),
		userRepo,
		userSettingsRepo,
		archivedUserRepo,
		emailVerificationRepo,
		usernameChangeHistoryRepo,
		passwordChangeHistoryRepo,
		fileStorageService,
		passwordService,
		emailService,
		logger,
	)

	// === Presenter層 ===
	authPresenter := presenter.NewAuthPresenter()
	pointPresenter := presenter.NewPointPresenter()
	friendPresenter := presenter.NewFriendPresenter()
	qrcodePresenter := presenter.NewQRCodePresenter()
	adminPresenter := presenter.NewAdminPresenter()
	userSettingsPresenter := presenter.NewUserSettingsPresenter()

	// === Controller層 ===
	authController := web.NewAuthController(authUC, authPresenter)
	pointController := web.NewPointController(pointTransferUC, pointPresenter)
	friendController := web.NewFriendController(friendshipUC, userRepo, friendPresenter)
	qrcodeController := web.NewQRCodeController(qrcodeUC, qrcodePresenter)
	adminController := web.NewAdminController(adminUC, adminPresenter)
	productController := web.NewProductController(productManagementUC, productExchangeUC, logger)
	categoryController := web.NewCategoryController(categoryUC, logger)
	userSettingsController := web.NewUserSettingsController(userSettingsUC, userSettingsPresenter)

	// === Middleware層 ===
	authMiddleware := middleware.NewAuthMiddleware(authUC)
	csrfMiddleware := middleware.NewCSRFMiddleware()

	// === Framework層 ===
	timeProvider := frameworksweb.NewSystemTimeProvider()
	router := frameworksweb.NewRouter(routerConfig, timeProvider)

	// ルート登録
	router.RegisterRoutes(
		authController,
		pointController,
		friendController,
		qrcodeController,
		adminController,
		productController,
		categoryController,
		userSettingsController,
		authMiddleware,
		csrfMiddleware,
	)

	logger.Info("AppContainer initialized successfully")
	logger.Info("All repositories, interactors, and controllers are ready")

	return &AppContainer{
		Router: router,
		DB:     db,
	}, nil
}

// Close はアプリケーションのリソースを解放
func (c *AppContainer) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

func main() {
	// 設定読み込み
	cfg := config.LoadConfig()

	// データベース設定
	dbConfig := &inframysql.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
		Env:      cfg.Server.Env,
	}

	// ルーター設定
	routerConfig := &frameworksweb.RouterConfig{
		Env:            cfg.Server.Env,
		AllowedOrigins: cfg.Security.AllowedOrigins,
	}

	// 依存性注入（手動DI）
	app, err := NewAppContainer(dbConfig, routerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	// サーバー起動
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Clean Architecture Server starting on %s (env: %s)", addr, cfg.Server.Env)
	log.Println("✅ Clean Architecture implementation is ready!")
	log.Println("📚 See CLEAN_ARCHITECTURE.md for architecture details")

	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
