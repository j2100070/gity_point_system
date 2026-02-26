package infrapostgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB はデータベース接続のインターフェース
// 内側のレイヤーが各ミドルウェアのI/Fを把握せずとも利用できる状態にする
type DB interface {
	GetDB() *gorm.DB
	Close() error
}

// PostgresDB はPostgreSQLの接続実装
type PostgresDB struct {
	db *gorm.DB
}

// Config はPostgreSQLの設定
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Env      string
}

// NewPostgresDB は新しいPostgresDBを作成
func NewPostgresDB(cfg *Config) (DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// GORM設定
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if cfg.Env == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	// PostgreSQL接続
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
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

	// トランザクション分離レベルをREPEATABLE READに設定
	// PostgreSQLのREPEATABLE READは、ファントムリードも防止する
	if err := db.Exec("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ").Error; err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// GetDB はGORMのDBインスタンスを取得
func (p *PostgresDB) GetDB() *gorm.DB {
	return p.db
}

// Close はデータベース接続を閉じる
func (p *PostgresDB) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
