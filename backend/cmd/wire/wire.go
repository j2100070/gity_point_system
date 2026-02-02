//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/infralogger"
	"github.com/gity/point-system/gateways/infra/inframysql"
	userrepo "github.com/gity/point-system/gateways/repository/user"
	frameworksweb "github.com/gity/point-system/frameworks/web"
	"github.com/gity/point-system/frameworks/web/middleware"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProvideGormDB は inframysql.DB から *gorm.DB を提供
func ProvideGormDB(db inframysql.DB) *gorm.DB {
	return db.GetDB()
}

// InitializeRouter はルーターを初期化（簡易版）
func InitializeRouter(dbConfig *inframysql.Config, routerConfig *frameworksweb.RouterConfig) (*frameworksweb.Router, error) {
	wire.Build(
		// Infra層
		inframysql.NewPostgresDB,
		ProvideGormDB,
		infralogger.NewLogger,

		// DataSource層
		dsmysqlimpl.NewUserDataSource,

		// Repository層
		userrepo.NewUserRepository,
		wire.Bind(new(interface{}), new(*userrepo.RepositoryImpl)),

		// Presenter層
		presenter.NewPointPresenter,

		// Framework層
		frameworksweb.NewSystemTimeProvider,
		frameworksweb.NewRouter,
	)

	return nil, nil
}
