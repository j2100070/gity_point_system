package config

import (
	"fmt"
	"os"
	"strings"
)

// Config はアプリケーション設定
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
}

// ServerConfig はサーバー設定
type ServerConfig struct {
	Port string
	Host string
	Env  string // development, production
}

// DatabaseConfig はデータベース設定
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// SecurityConfig はセキュリティ設定
type SecurityConfig struct {
	AllowedOrigins []string // CORS許可オリジン
	SessionSecret  string   // セッション暗号化キー
}

// LoadConfig は設定をロード
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "point_system"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Security: SecurityConfig{
			AllowedOrigins: getAllowedOrigins(),
			SessionSecret:  getEnv("SESSION_SECRET", "change-this-in-production-very-secret-key-32bytes"),
		},
	}
}

// GetDSN はPostgreSQL接続文字列を返す
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// getEnv は環境変数を取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getAllowedOrigins はALLOWED_ORIGINS環境変数からオリジンリストを取得
func getAllowedOrigins() []string {
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	origins := strings.Split(originsStr, ",")

	// 空白を削除
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}
