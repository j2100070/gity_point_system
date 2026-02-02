package main

import (
	"fmt"
	"log"

	"github.com/gity/point-system/config"
	"github.com/gity/point-system/gateways/infra/inframysql"
	frameworksweb "github.com/gity/point-system/frameworks/web"
)

func mainClean() {
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
	log.Printf("Server starting on %s (env: %s)", addr, cfg.Server.Env)
	log.Println("Clean Architecture implementation is ready!")

	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
