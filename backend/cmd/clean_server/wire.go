//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gity/point-system/config"
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/entities"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/frameworks/web/middleware"
	"github.com/gity/point-system/gateways/infra/infraemail"
	"github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/gateways/infra/infrastorage"
	"github.com/gity/point-system/usecases/service"
	"github.com/google/wire"
)

// InitializeApp は Wire を使ってアプリケーションの依存関係を自動注入する
func InitializeApp(cfg *config.Config) (*AppContainer, error) {
	wire.Build(
		// Config providers
		ProvideDBConfig,
		ProvideRouterConfig,
		ProvideFileStorageService,
		ProvideEmailService,

		// レイヤー別 ProviderSet
		InfraSet,
		DataSourceSet,
		RepositorySet,
		ServiceSet,
		InteractorSet,
		PresenterSet,
		ControllerSet,
		MiddlewareSet,
		FrameworkSet,

		// Router
		ProvideRouter,

		// AppContainer
		wire.Struct(new(AppContainer), "*"),
	)
	return nil, nil
}

// ========================================
// Config Providers
// ========================================

func ProvideDBConfig(cfg *config.Config) *infrapostgres.Config {
	return &infrapostgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
		Env:      cfg.Server.Env,
	}
}

func ProvideRouterConfig(cfg *config.Config) *frameworksweb.RouterConfig {
	return &frameworksweb.RouterConfig{
		Env:             cfg.Server.Env,
		AllowedOrigins:  cfg.Security.AllowedOrigins,
		MaxUploadSizeMB: cfg.Server.MaxUploadSizeMB,
	}
}

func ProvideFileStorageService() (service.FileStorageService, error) {
	return infrastorage.NewLocalStorage(&infrastorage.Config{
		BaseDir:   "./uploads/avatars",
		BaseURL:   "/uploads/avatars",
		MaxSizeMB: 20,
	})
}

func ProvideEmailService(logger entities.Logger) service.EmailService {
	return infraemail.NewConsoleEmailService(logger)
}

// ========================================
// Router Provider
// ========================================

func ProvideRouter(
	cfg *frameworksweb.RouterConfig,
	tp frameworksweb.TimeProvider,
	auth *web.AuthController,
	point *web.PointController,
	friend *web.FriendController,
	qrcode *web.QRCodeController,
	transferReq *web.TransferRequestController,
	dailyBonus *web.DailyBonusController,
	admin *web.AdminController,
	product *web.ProductController,
	category *web.CategoryController,
	settings *web.UserSettingsController,
	authMW *middleware.AuthMiddleware,
	csrfMW *middleware.CSRFMiddleware,
) *frameworksweb.Router {
	r := frameworksweb.NewRouter(cfg, tp)
	r.RegisterRoutes(
		auth, point, friend, qrcode, transferReq,
		dailyBonus, admin, product, category, settings,
		authMW, csrfMW,
	)
	return r
}
