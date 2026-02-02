package main

import (
	"log"

	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/infralogger"
	"github.com/gity/point-system/gateways/infra/inframysql"
	userrepo "github.com/gity/point-system/gateways/repository/user"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/frameworks/web/middleware"
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

	// === DataSource層 ===
	userDS := dsmysqlimpl.NewUserDataSource(db)

	// === Repository層 ===
	userRepo := userrepo.NewUserRepository(userDS, logger)

	// === Interactor層 ===
	// TODO: 他のRepositoryが実装されたら追加
	// pointTransferUC := interactor.NewPointTransferInteractor(
	// 	db.GetDB(),
	// 	userRepo,
	// 	transactionRepo,
	// 	idempotencyRepo,
	// 	friendshipRepo,
	// 	logger,
	// )

	// authUC := interactor.NewAuthInteractor(
	// 	userRepo,
	// 	sessionRepo,
	// 	logger,
	// )

	// === Presenter層 ===
	pointPresenter := presenter.NewPointPresenter()

	// === Controller層 ===
	// pointController := web.NewPointController(pointTransferUC, pointPresenter)

	// === Middleware層 ===
	// authMiddleware := middleware.NewAuthMiddleware(authUC)
	csrfMiddleware := middleware.NewCSRFMiddleware()

	// === Framework層 ===
	timeProvider := frameworksweb.NewSystemTimeProvider()
	router := frameworksweb.NewRouter(routerConfig, timeProvider)

	// ルート登録
	// router.RegisterRoutes(pointController, authMiddleware, csrfMiddleware)

	// 一時的にログ出力（TODO: 削除）
	log.Println("AppContainer initialized successfully")
	log.Printf("PointPresenter: %v", pointPresenter)
	log.Printf("CSRFMiddleware: %v", csrfMiddleware)
	log.Printf("UserRepo: %v", userRepo)

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
