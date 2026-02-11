package infrastorage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gity/point-system/usecases/service"
)

// LocalStorage はローカルファイルシステムを使用したストレージ実装
type LocalStorage struct {
	baseDir    string // 保存先のベースディレクトリ
	baseURL    string // アクセス用のベースURL
	maxSizeMB  int64  // 最大ファイルサイズ（MB）
	allowedExt map[string]bool
}

// Config はLocalStorageの設定
type Config struct {
	BaseDir    string   // 例: "./uploads/avatars"
	BaseURL    string   // 例: "/uploads/avatars"
	MaxSizeMB  int64    // 例: 20 (20MB)
	AllowedExt []string // 例: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
}

// NewLocalStorage は新しいLocalStorageを作成
func NewLocalStorage(cfg *Config) (service.FileStorageService, error) {
	if cfg.BaseDir == "" {
		return nil, errors.New("base directory is required")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("base URL is required")
	}
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 20 // デフォルト20MB
	}

	// 許可する拡張子のマップを作成
	allowedExt := make(map[string]bool)
	if len(cfg.AllowedExt) == 0 {
		// デフォルトで許可する拡張子
		cfg.AllowedExt = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	}
	for _, ext := range cfg.AllowedExt {
		allowedExt[strings.ToLower(ext)] = true
	}

	// ベースディレクトリが存在しない場合は作成
	if err := os.MkdirAll(cfg.BaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		baseDir:    cfg.BaseDir,
		baseURL:    cfg.BaseURL,
		maxSizeMB:  cfg.MaxSizeMB,
		allowedExt: allowedExt,
	}, nil
}

// SaveAvatar はアバター画像を保存
func (s *LocalStorage) SaveAvatar(userID string, fileName string, file io.Reader, fileSize int64) (string, error) {
	// ファイルサイズチェック
	maxBytes := s.maxSizeMB * 1024 * 1024
	if fileSize > maxBytes {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %dMB", s.maxSizeMB)
	}

	// 拡張子チェック
	ext := strings.ToLower(filepath.Ext(fileName))
	if !s.allowedExt[ext] {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	// ユーザーごとのディレクトリを作成
	userDir := filepath.Join(s.baseDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// ファイル名を生成（タイムスタンプ + ハッシュ）
	timestamp := time.Now().Unix()
	hash := s.generateFileHash(userID, fileName, timestamp)
	newFileName := fmt.Sprintf("%d_%s%s", timestamp, hash[:12], ext)
	filePath := filepath.Join(userDir, newFileName)

	// ファイルを保存
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// ファイルサイズを制限しながらコピー
	limitedReader := io.LimitReader(file, maxBytes)
	written, err := io.Copy(outFile, limitedReader)
	if err != nil {
		// エラー時はファイルを削除
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// ファイルサイズの再確認
	if written != fileSize {
		os.Remove(filePath)
		return "", errors.New("file size mismatch")
	}

	// 相対パスを返す（例: "userID/timestamp_hash.jpg"）
	relativePath := filepath.Join(userID, newFileName)
	return relativePath, nil
}

// DeleteAvatar はアバター画像を削除
func (s *LocalStorage) DeleteAvatar(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}

	// セキュリティチェック: パストラバーサル攻撃を防ぐ
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return errors.New("invalid file path")
	}

	fullPath := filepath.Join(s.baseDir, cleanPath)

	// ファイルが存在するか確認
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// ファイルが存在しない場合はエラーとしない（冪等性）
		return nil
	}

	// ファイルを削除
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetAvatarURL はアバター画像のURLを取得
func (s *LocalStorage) GetAvatarURL(filePath string) string {
	if filePath == "" {
		return ""
	}

	// パスの正規化
	cleanPath := filepath.Clean(filePath)

	// URLを構築（スラッシュ区切りに変換）
	urlPath := filepath.ToSlash(cleanPath)
	return fmt.Sprintf("%s/%s", s.baseURL, urlPath)
}

// generateFileHash はファイル名のハッシュを生成
func (s *LocalStorage) generateFileHash(userID, fileName string, timestamp int64) string {
	data := fmt.Sprintf("%s:%s:%d", userID, fileName, timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
