package main

import (
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/controllers/web/presenter"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/frameworks/web/middleware"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/gity/point-system/gateways/infra/infralogger"
	"github.com/gity/point-system/gateways/infra/infrapassword"
	"github.com/gity/point-system/gateways/infra/infrapostgres"
	categoryrepo "github.com/gity/point-system/gateways/repository/category"
	dailybonusrepo "github.com/gity/point-system/gateways/repository/daily_bonus"
	dsmysql "github.com/gity/point-system/gateways/repository/datasource/dsmysql"
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
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/wire"
)

// ========================================
// Infra ProviderSet
// ========================================

var InfraSet = wire.NewSet(
	infrapostgres.NewPostgresDB,
	infralogger.NewLogger,
	ProvideGormTransactionManager,
	wire.Bind(new(repository.TransactionManager), new(*infrapostgres.GormTransactionManager)),
)

// ProvideGormTransactionManager は DB から TransactionManager を作成
func ProvideGormTransactionManager(db infrapostgres.DB) *infrapostgres.GormTransactionManager {
	return infrapostgres.NewGormTransactionManager(db.GetDB())
}

// ========================================
// DataSource ProviderSet
// ========================================

var DataSourceSet = wire.NewSet(
	dspostgresimpl.NewUserDataSource,
	dspostgresimpl.NewTransactionDataSource,
	dspostgresimpl.NewIdempotencyKeyDataSource,
	dspostgresimpl.NewSessionDataSource,
	dspostgresimpl.NewFriendshipDataSource,
	dspostgresimpl.NewQRCodeDataSource,
	dspostgresimpl.NewTransferRequestDataSource,
	dspostgresimpl.NewDailyBonusDataSource,
	dspostgresimpl.NewProductDataSource,
	dspostgresimpl.NewProductExchangeDataSource,
	dspostgresimpl.NewCategoryDataSource,
	dspostgresimpl.NewArchivedUserDataSource,
	dspostgresimpl.NewEmailVerificationDataSource,
	dspostgresimpl.NewUsernameChangeHistoryDataSource,
	dspostgresimpl.NewPasswordChangeHistoryDataSource,
	dspostgresimpl.NewSystemSettingsDataSource,
	dspostgresimpl.NewPointBatchDataSource,
	dspostgresimpl.NewLotteryTierDataSource,
	dspostgresimpl.NewAnalyticsDataSource,

	// concrete → interface bindings (DataSource constructors that return *Impl instead of interface)
	wire.Bind(new(dsmysql.ArchivedUserDataSource), new(*dspostgresimpl.ArchivedUserDataSourceImpl)),
	wire.Bind(new(dsmysql.EmailVerificationDataSource), new(*dspostgresimpl.EmailVerificationDataSourceImpl)),
	wire.Bind(new(dsmysql.UsernameChangeHistoryDataSource), new(*dspostgresimpl.UsernameChangeHistoryDataSourceImpl)),
	wire.Bind(new(dsmysql.PasswordChangeHistoryDataSource), new(*dspostgresimpl.PasswordChangeHistoryDataSourceImpl)),

	// AnalyticsDataSource → repository.AnalyticsRepository
	// AnalyticsDataSourceは dsmysql.AnalyticsDataSource を返すが、AdminInteractorは repository.AnalyticsRepository を要求
	wire.Bind(new(repository.AnalyticsRepository), new(dsmysql.AnalyticsDataSource)),
)

// ========================================
// Repository ProviderSet
// ========================================

var RepositorySet = wire.NewSet(
	userrepo.NewUserRepository,
	transactionrepo.NewTransactionRepository,
	transactionrepo.NewIdempotencyKeyRepository,
	sessionrepo.NewSessionRepository,
	friendshiprepo.NewFriendshipRepository,
	qrcoderepo.NewQRCodeRepository,
	transferrequestrepo.NewTransferRequestRepository,
	dailybonusrepo.NewDailyBonusRepository,
	productrepo.NewProductRepository,
	productrepo.NewProductExchangeRepository,
	categoryrepo.NewCategoryRepository,
	usersettingsrepo.NewUserSettingsRepository,
	usersettingsrepo.NewArchivedUserRepository,
	usersettingsrepo.NewEmailVerificationRepository,
	usersettingsrepo.NewUsernameChangeHistoryRepository,
	usersettingsrepo.NewPasswordChangeHistoryRepository,
	systemsettingsrepo.NewSystemSettingsRepository,
	pointbatchrepo.NewPointBatchRepository,
	lotterytierrepo.NewLotteryTierRepository,

	// concrete → interface bindings
	wire.Bind(new(repository.DailyBonusRepository), new(*dailybonusrepo.DailyBonusRepositoryImpl)),
	wire.Bind(new(repository.SystemSettingsRepository), new(*systemsettingsrepo.SystemSettingsRepositoryImpl)),
	wire.Bind(new(repository.PointBatchRepository), new(*pointbatchrepo.PointBatchRepositoryImpl)),
	wire.Bind(new(repository.LotteryTierRepository), new(*lotterytierrepo.LotteryTierRepositoryImpl)),
)

// ========================================
// Service ProviderSet
// ========================================

var ServiceSet = wire.NewSet(
	infrapassword.NewBcryptPasswordService,
)

// ========================================
// Interactor ProviderSet
// ========================================

var InteractorSet = wire.NewSet(
	interactor.NewAuthInteractor,
	interactor.NewPointTransferInteractor,
	interactor.NewFriendshipInteractor,
	interactor.NewTransferRequestInteractor,
	interactor.NewDailyBonusInteractor,
	interactor.NewQRCodeInteractor,
	interactor.NewAdminInteractor,
	interactor.NewProductManagementInteractor,
	interactor.NewProductExchangeInteractor,
	interactor.NewCategoryManagementInteractor,
	interactor.NewUserQueryInteractor,
	interactor.NewUserSettingsInteractor,

	// concrete → interface bindings
	wire.Bind(new(inputport.PointTransferInputPort), new(*interactor.PointTransferInteractor)),
	wire.Bind(new(inputport.DailyBonusInputPort), new(*interactor.DailyBonusInteractor)),
	wire.Bind(new(inputport.ProductExchangeInputPort), new(*interactor.ProductExchangeInteractor)),
)

// ========================================
// Presenter ProviderSet
// ========================================

var PresenterSet = wire.NewSet(
	presenter.NewAuthPresenter,
	presenter.NewPointPresenter,
	presenter.NewFriendPresenter,
	presenter.NewQRCodePresenter,
	presenter.NewTransferRequestPresenter,
	presenter.NewDailyBonusPresenter,
	presenter.NewAdminPresenter,
	presenter.NewUserSettingsPresenter,
)

// ========================================
// Controller ProviderSet
// ========================================

var ControllerSet = wire.NewSet(
	web.NewAuthController,
	web.NewPointController,
	web.NewFriendController,
	web.NewQRCodeController,
	web.NewTransferRequestController,
	web.NewDailyBonusController,
	web.NewAdminController,
	web.NewProductController,
	web.NewCategoryController,
	web.NewUserSettingsController,
)

// ========================================
// Middleware ProviderSet
// ========================================

var MiddlewareSet = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewCSRFMiddleware,
)

// ========================================
// Router & Framework ProviderSet
// ========================================

var FrameworkSet = wire.NewSet(
	frameworksweb.NewSystemTimeProvider,
)
