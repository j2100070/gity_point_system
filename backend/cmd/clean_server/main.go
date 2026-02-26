package main

import (
	"fmt"
	"log"

	"github.com/gity/point-system/config"
	"github.com/gity/point-system/entities"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/gity/point-system/gateways/infra"
	"github.com/gity/point-system/gateways/infra/infraakerun"
	"github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
)

// AppContainer はアプリケーションの依存関係を管理
// Wire が自動注入するフィールド
type AppContainer struct {
	Router *frameworksweb.Router
	DB     infrapostgres.DB

	// Workers 構築に必要な依存を Wire から受け取る
	DailyBonusUC    *interactor.DailyBonusInteractor
	PointBatchRepo  repository.PointBatchRepository
	UserRepo        repository.UserRepository
	TransactionRepo repository.TransactionRepository
	TxManager       repository.TransactionManager
	Logger          entities.Logger
	TimeProvider    frameworksweb.TimeProvider
}

func main() {
	cfg := config.LoadConfig()

	// Wire DI
	app, err := InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer func() {
		if app.DB != nil {
			app.DB.Close()
		}
	}()

	// AutoMigrate（新規テーブルのみ）
	if err := app.DB.GetDB().AutoMigrate(
		&dspostgresimpl.CategoryModel{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	// Workers（Wire 外で構築）
	startWorkers(cfg, app)

	// サーバー起動
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s (env: %s)", addr, cfg.Server.Env)

	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startWorkers(cfg *config.Config, app *AppContainer) {
	// Akerun Worker
	akerunClient := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
		AccessToken:    cfg.Akerun.AccessToken,
		OrganizationID: cfg.Akerun.OrganizationID,
	})
	akerunWorker := infraakerun.NewAkerunWorker(
		akerunClient, app.DailyBonusUC, app.TimeProvider, app.Logger,
	)
	akerunWorker.Start()

	// Point Expiry Worker
	pointExpiryWorker := infra.NewPointExpiryWorker(
		app.PointBatchRepo, app.UserRepo, app.TransactionRepo,
		app.TxManager, app.Logger,
	)
	pointExpiryWorker.Start()

	app.Logger.Info("All workers started")
}
