package service

import (
	"io"
)

// FileStorageService はファイル保存サービスのインターフェース
type FileStorageService interface {
	// SaveAvatar はアバター画像を保存し、保存先のパス（URL）を返す
	SaveAvatar(userID string, fileName string, file io.Reader, fileSize int64) (string, error)

	// DeleteAvatar はアバター画像を削除
	DeleteAvatar(filePath string) error

	// GetAvatarURL はアバター画像のURLを取得
	GetAvatarURL(filePath string) string
}
