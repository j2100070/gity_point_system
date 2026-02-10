package infrastorage_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gity/point-system/gateways/infra/infrastorage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// テスト用の一時ディレクトリを作成
func setupTestStorage(t *testing.T) (*infrastorage.Config, string) {
	tempDir := filepath.Join(os.TempDir(), "test_avatars_"+uuid.New().String())
	cfg := &infrastorage.Config{
		BaseDir:    tempDir,
		BaseURL:    "/uploads/avatars",
		MaxSizeMB:  5,
		AllowedExt: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
	}
	return cfg, tempDir
}

// テスト終了時に一時ディレクトリを削除
func cleanupTestStorage(t *testing.T, tempDir string) {
	os.RemoveAll(tempDir)
}

// ========================================
// LocalStorage Tests
// ========================================

func TestNewLocalStorage(t *testing.T) {
	t.Run("正常にストレージを作成", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)
		assert.NotNil(t, storage)

		// ベースディレクトリが作成されていることを確認
		_, err = os.Stat(tempDir)
		assert.NoError(t, err)
	})

	t.Run("ベースディレクトリが空の場合はエラー", func(t *testing.T) {
		cfg := &infrastorage.Config{
			BaseDir: "",
			BaseURL: "/uploads/avatars",
		}

		_, err := infrastorage.NewLocalStorage(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base directory is required")
	})

	t.Run("ベースURLが空の場合はエラー", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)
		cfg.BaseURL = ""

		_, err := infrastorage.NewLocalStorage(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base URL is required")
	})
}

func TestLocalStorage_SaveAvatar(t *testing.T) {
	t.Run("正常にアバターを保存", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		userID := uuid.New().String()
		fileName := "avatar.jpg"
		fileContent := []byte("fake image content")
		file := bytes.NewReader(fileContent)

		filePath, err := storage.SaveAvatar(userID, fileName, file, int64(len(fileContent)))
		require.NoError(t, err)
		assert.NotEmpty(t, filePath)
		assert.Contains(t, filePath, userID)
		assert.True(t, strings.HasSuffix(filePath, ".jpg"))

		// ファイルが実際に保存されているか確認
		fullPath := filepath.Join(tempDir, filePath)
		_, err = os.Stat(fullPath)
		assert.NoError(t, err)

		// ファイル内容を確認
		savedContent, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, fileContent, savedContent)
	})

	t.Run("ファイルサイズが制限を超える場合はエラー", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)
		cfg.MaxSizeMB = 1 // 1MBに制限

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		userID := uuid.New().String()
		fileName := "large_avatar.jpg"
		// 2MBのダミーデータ
		fileSize := int64(2 * 1024 * 1024)
		file := bytes.NewReader(make([]byte, fileSize))

		_, err = storage.SaveAvatar(userID, fileName, file, fileSize)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum allowed size")
	})

	t.Run("許可されていない拡張子の場合はエラー", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		userID := uuid.New().String()
		fileName := "avatar.exe" // 許可されていない拡張子
		fileContent := []byte("fake content")
		file := bytes.NewReader(fileContent)

		_, err = storage.SaveAvatar(userID, fileName, file, int64(len(fileContent)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not allowed")
	})

	t.Run("複数のファイルを保存できる", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		userID := uuid.New().String()

		// 1つ目のファイル
		file1 := bytes.NewReader([]byte("image1"))
		path1, err := storage.SaveAvatar(userID, "avatar1.jpg", file1, 6)
		require.NoError(t, err)

		// 2つ目のファイル
		file2 := bytes.NewReader([]byte("image2"))
		path2, err := storage.SaveAvatar(userID, "avatar2.png", file2, 6)
		require.NoError(t, err)

		// パスが異なることを確認
		assert.NotEqual(t, path1, path2)

		// 両方のファイルが存在することを確認
		_, err = os.Stat(filepath.Join(tempDir, path1))
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(tempDir, path2))
		assert.NoError(t, err)
	})
}

func TestLocalStorage_DeleteAvatar(t *testing.T) {
	t.Run("正常にアバターを削除", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		// ファイルを保存
		userID := uuid.New().String()
		fileContent := []byte("test image")
		file := bytes.NewReader(fileContent)
		filePath, err := storage.SaveAvatar(userID, "avatar.jpg", file, int64(len(fileContent)))
		require.NoError(t, err)

		// ファイルが存在することを確認
		fullPath := filepath.Join(tempDir, filePath)
		_, err = os.Stat(fullPath)
		require.NoError(t, err)

		// ファイルを削除
		err = storage.DeleteAvatar(filePath)
		require.NoError(t, err)

		// ファイルが削除されていることを確認
		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("存在しないファイルの削除は成功（冪等性）", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		err = storage.DeleteAvatar("user123/nonexistent.jpg")
		assert.NoError(t, err) // エラーにならない
	})

	t.Run("パストラバーサル攻撃を防ぐ", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		err = storage.DeleteAvatar("../../etc/passwd")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid file path")
	})

	t.Run("空のパスはエラー", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		err = storage.DeleteAvatar("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file path is empty")
	})
}

func TestLocalStorage_GetAvatarURL(t *testing.T) {
	t.Run("正しいURLを生成", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		filePath := "user123/1234567890_abcdef123456.jpg"
		url := storage.GetAvatarURL(filePath)

		assert.Equal(t, "/uploads/avatars/user123/1234567890_abcdef123456.jpg", url)
	})

	t.Run("空のパスは空文字列を返す", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		url := storage.GetAvatarURL("")
		assert.Equal(t, "", url)
	})

	t.Run("パスの正規化", func(t *testing.T) {
		cfg, tempDir := setupTestStorage(t)
		defer cleanupTestStorage(t, tempDir)

		storage, err := infrastorage.NewLocalStorage(cfg)
		require.NoError(t, err)

		// 相対パス記号を含むパスも正規化される
		filePath := "user123/./subfolder/../avatar.jpg"
		url := storage.GetAvatarURL(filePath)

		// パスが正規化されていることを確認
		assert.Contains(t, url, "user123/avatar.jpg")
		assert.NotContains(t, url, "..")
	})
}

// ========================================
// 統合的なシナリオテスト
// ========================================

func TestLocalStorage_Scenario_SaveUpdateDelete(t *testing.T) {
	cfg, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tempDir)

	storage, err := infrastorage.NewLocalStorage(cfg)
	require.NoError(t, err)

	userID := uuid.New().String()

	// 1. 最初のアバターを保存
	file1 := bytes.NewReader([]byte("first avatar"))
	path1, err := storage.SaveAvatar(userID, "avatar1.jpg", file1, 12)
	require.NoError(t, err)

	// URLを取得
	url1 := storage.GetAvatarURL(path1)
	assert.Contains(t, url1, userID)

	// 2. 新しいアバターを保存（更新をシミュレート）
	file2 := bytes.NewReader([]byte("second avatar"))
	path2, err := storage.SaveAvatar(userID, "avatar2.png", file2, 13)
	require.NoError(t, err)

	// パスが異なることを確認
	assert.NotEqual(t, path1, path2)

	// 3. 古いアバターを削除
	err = storage.DeleteAvatar(path1)
	require.NoError(t, err)

	// 古いファイルが削除されていることを確認
	_, err = os.Stat(filepath.Join(tempDir, path1))
	assert.True(t, os.IsNotExist(err))

	// 新しいファイルは存在することを確認
	_, err = os.Stat(filepath.Join(tempDir, path2))
	assert.NoError(t, err)

	// 4. 最終的にアバターを削除（アカウント削除をシミュレート）
	err = storage.DeleteAvatar(path2)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tempDir, path2))
	assert.True(t, os.IsNotExist(err))
}
