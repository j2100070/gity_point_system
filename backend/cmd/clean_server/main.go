package main

import (
	"fmt"
	"log"

	"github.com/gity/point-system/config"
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/controllers/web/presenter"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/frameworks/web/middleware"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra"
	"github.com/gity/point-system/gateways/infra/infraakerun"
	"github.com/gity/point-system/gateways/infra/infraemail"
	"github.com/gity/point-system/gateways/infra/infralogger"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/infra/infrapassword"
	"github.com/gity/point-system/gateways/infra/infrastorage"
	categoryrepo "github.com/gity/point-system/gateways/repository/category"
	dailybonusrepo "github.com/gity/point-system/gateways/repository/daily_bonus"
	friendshiprepo "github.com/gity/point-system/gateways/repository/friendship"
	lotterytierrepo "github.com/gity/point-system/gateways/repository/lottery_tier"
	pointbatchrepo "github.com/gity/point-system/gateways/repository/point_batch"
	productrepo "github.com/gity/point-system/gateways/repository/product"
	qrcoderepo "github.com/gity/point-system/gateways/repository/qrcode"
	sessionrepo "github.com/gity/point-system/gateways/repository/session"
	systemsettingsrepo "github.com/gity/point-system/gateways/repository/system_settings"
	transactionrepo "github.com/gity/point-system/gateways/repository/transaction"
	transferrequestrepo "github.com/gity/point-system/gateways/repository/transfer_request"
	userrepo "github.com/gity/point-system/gateways/repository/user"
	usersettingsrepo "github.com/gity/point-system/gateways/repository/user_settings"
	"github.com/gity/point-system/usecases/interactor"
)

// AppContainer はアプリケーションの依存関係を管理
type AppContainer struct {
	Router            *frameworksweb.Router
	DB                inframysql.DB
	akerunWorker      *infraakerun.AkerunWorker
	pointExpiryWorker *infra.PointExpiryWorker
}

// NewAppContainer は新しいAppContainerを作成（手動DI）
func NewAppContainer(cfg *config.Config) (*AppContainer, error) {
	// === Infra層 ===
	dbConfig := &inframysql.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
		Env:      cfg.Server.Env,
	}

	routerConfig := &frameworksweb.RouterConfig{
		Env:             cfg.Server.Env,
		AllowedOrigins:  cfg.Security.AllowedOrigins,
		MaxUploadSizeMB: cfg.Server.MaxUploadSizeMB,
	}

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
	transferRequestDS := dsmysqlimpl.NewTransferRequestDataSource(db)
	dailyBonusDS := dsmysqlimpl.NewDailyBonusDataSource(db)
	productDS := dsmysqlimpl.NewProductDataSource(db)
	productExchangeDS := dsmysqlimpl.NewProductExchangeDataSource(db)
	categoryDS := dsmysqlimpl.NewCategoryDataSource(db)
	archivedUserDS := dsmysqlimpl.NewArchivedUserDataSource(db)
	emailVerificationDS := dsmysqlimpl.NewEmailVerificationDataSource(db)
	usernameChangeHistoryDS := dsmysqlimpl.NewUsernameChangeHistoryDataSource(db)
	passwordChangeHistoryDS := dsmysqlimpl.NewPasswordChangeHistoryDataSource(db)
	systemSettingsDS := dsmysqlimpl.NewSystemSettingsDataSource(db)
	pointBatchDS := dsmysqlimpl.NewPointBatchDataSource(db)
	lotteryTierDS := dsmysqlimpl.NewLotteryTierDataSource(db)

	// === Repository層 ===
	userRepo := userrepo.NewUserRepository(userDS, logger)
	transactionRepo := transactionrepo.NewTransactionRepository(transactionDS, logger)
	idempotencyRepo := transactionrepo.NewIdempotencyKeyRepository(idempotencyDS, logger)
	sessionRepo := sessionrepo.NewSessionRepository(sessionDS, logger)
	friendshipRepo := friendshiprepo.NewFriendshipRepository(friendshipDS, logger)
	qrcodeRepo := qrcoderepo.NewQRCodeRepository(qrcodeDS, logger)
	transferRequestRepo := transferrequestrepo.NewTransferRequestRepository(transferRequestDS, logger)
	dailyBonusRepo := dailybonusrepo.NewDailyBonusRepository(dailyBonusDS)
	productRepo := productrepo.NewProductRepository(productDS, logger)
	productExchangeRepo := productrepo.NewProductExchangeRepository(productExchangeDS, logger)
	categoryRepo := categoryrepo.NewCategoryRepository(categoryDS, logger)
	userSettingsRepo := usersettingsrepo.NewUserSettingsRepository(userDS, logger)
	archivedUserRepo := usersettingsrepo.NewArchivedUserRepository(archivedUserDS, logger)
	emailVerificationRepo := usersettingsrepo.NewEmailVerificationRepository(emailVerificationDS, logger)
	usernameChangeHistoryRepo := usersettingsrepo.NewUsernameChangeHistoryRepository(usernameChangeHistoryDS, logger)
	passwordChangeHistoryRepo := usersettingsrepo.NewPasswordChangeHistoryRepository(passwordChangeHistoryDS, logger)
	systemSettingsRepo := systemsettingsrepo.NewSystemSettingsRepository(systemSettingsDS)
	pointBatchRepo := pointbatchrepo.NewPointBatchRepository(pointBatchDS)
	lotteryTierRepo := lotterytierrepo.NewLotteryTierRepository(lotteryTierDS)

	// === Service層 ===
	passwordService := infrapassword.NewBcryptPasswordService()

	// === Interactor層 ===
	authUC := interactor.NewAuthInteractor(
		userRepo,
		sessionRepo,
		passwordService,
		logger,
	)

	// TransactionManagerを作成（他のInteractorより先に作成）
	txManager := inframysql.NewGormTransactionManager(db.GetDB())

	pointTransferUC := interactor.NewPointTransferInteractor(
		txManager,
		userRepo,
		transactionRepo,
		idempotencyRepo,
		friendshipRepo,
		pointBatchRepo,
		logger,
	)

	friendshipUC := interactor.NewFriendshipInteractor(
		friendshipRepo,
		userRepo,
		logger,
	)

	transferRequestUC := interactor.NewTransferRequestInteractor(
		transferRequestRepo,
		userRepo,
		pointTransferUC,
		logger,
	)

	dailyBonusUC := interactor.NewDailyBonusInteractor(
		dailyBonusRepo,
		userRepo,
		transactionRepo,
		txManager,
		systemSettingsRepo,
		pointBatchRepo,
		lotteryTierRepo,
		logger,
	)

	qrcodeUC := interactor.NewQRCodeInteractor(
		qrcodeRepo,
		pointTransferUC,
		logger,
	)

	adminUC := interactor.NewAdminInteractor(
		txManager,
		userRepo,
		transactionRepo,
		idempotencyRepo,
		pointBatchRepo,
		dsmysqlimpl.NewAnalyticsDataSource(db),
		logger,
	)

	productManagementUC := interactor.NewProductManagementInteractor(
		productRepo,
		logger,
	)

	productExchangeUC := interactor.NewProductExchangeInteractor(
		txManager,
		productRepo,
		productExchangeRepo,
		userRepo,
		transactionRepo,
		pointBatchRepo,
		logger,
	)

	categoryUC := interactor.NewCategoryManagementInteractor(
		categoryRepo,
		logger,
	)

	userQueryUC := interactor.NewUserQueryInteractor(
		userRepo,
		logger,
	)

	// === Additional Service層 ===
	fileStorageService, err := infrastorage.NewLocalStorage(&infrastorage.Config{
		BaseDir:   "./uploads/avatars",
		BaseURL:   "/uploads/avatars",
		MaxSizeMB: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file storage service: %w", err)
	}
	emailService := infraemail.NewConsoleEmailService(logger)

	userSettingsUC := interactor.NewUserSettingsInteractor(
		txManager,
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
	transferRequestPresenter := presenter.NewTransferRequestPresenter()
	dailyBonusPresenter := presenter.NewDailyBonusPresenter()
	adminPresenter := presenter.NewAdminPresenter()
	userSettingsPresenter := presenter.NewUserSettingsPresenter()

	// === Controller層 ===
	authController := web.NewAuthController(authUC, authPresenter)
	pointController := web.NewPointController(pointTransferUC, pointPresenter)
	friendController := web.NewFriendController(friendshipUC, userQueryUC, friendPresenter)
	qrcodeController := web.NewQRCodeController(qrcodeUC, qrcodePresenter)
	transferRequestController := web.NewTransferRequestController(transferRequestUC, userQueryUC, transferRequestPresenter)
	dailyBonusController := web.NewDailyBonusController(dailyBonusUC, dailyBonusPresenter)
	adminController := web.NewAdminController(adminUC, adminPresenter)
	productController := web.NewProductController(productManagementUC, productExchangeUC, logger)
	categoryController := web.NewCategoryController(categoryUC, logger)
	userSettingsController := web.NewUserSettingsController(userSettingsUC, userSettingsPresenter)

	// === Middleware層 ===
	authMiddleware := middleware.NewAuthMiddleware(authUC)
	csrfMiddleware := middleware.NewCSRFMiddleware()

	// === Framework層 ===
	frameworkTimeProvider := frameworksweb.NewSystemTimeProvider()
	router := frameworksweb.NewRouter(routerConfig, frameworkTimeProvider)

	// ルート登録
	router.RegisterRoutes(
		authController,
		pointController,
		friendController,
		qrcodeController,
		transferRequestController,
		dailyBonusController,
		adminController,
		productController,
		categoryController,
		userSettingsController,
		authMiddleware,
		csrfMiddleware,
	)

	// === Akerun Worker層 ===
	akerunClient := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
		AccessToken:    cfg.Akerun.AccessToken,
		OrganizationID: cfg.Akerun.OrganizationID,
	})

	akerunWorker := infraakerun.NewAkerunWorker(
		akerunClient,
		dailyBonusUC,
		frameworkTimeProvider,
		logger,
	)

	pointExpiryWorker := infra.NewPointExpiryWorker(
		pointBatchRepo,
		userRepo,
		transactionRepo,
		txManager,
		logger,
	)

	logger.Info("AppContainer initialized successfully")
	logger.Info("All repositories, interactors, and controllers are ready")

	return &AppContainer{
		Router:            router,
		DB:                db,
		akerunWorker:      akerunWorker,
		pointExpiryWorker: pointExpiryWorker,
	}, nil
}

// Close はアプリケーションのリソースを解放
func (c *AppContainer) Close() error {
	if c.akerunWorker != nil {
		c.akerunWorker.Stop()
	}
	if c.pointExpiryWorker != nil {
		c.pointExpiryWorker.Stop()
	}
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// StartWorkers はバックグラウンドワーカーを開始
func (c *AppContainer) StartWorkers() {
	if c.akerunWorker != nil {
		c.akerunWorker.Start()
	}
	if c.pointExpiryWorker != nil {
		c.pointExpiryWorker.Start()
	}
}

func main() {
	// 設定読み込み
	cfg := config.LoadConfig()

	// 依存性注入（手動DI）
	app, err := NewAppContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	// バックグラウンドワーカー開始
	app.StartWorkers()

	// サーバー起動
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Clean Architecture Server starting on %s (env: %s)", addr, cfg.Server.Env)
	log.Println("✅ Clean Architecture implementation is ready!")
	log.Println("📚 See CLEAN_ARCHITECTURE.md for architecture details")

	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
