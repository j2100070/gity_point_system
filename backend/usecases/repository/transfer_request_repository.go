package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransferRequestRepository は送金リクエストのリポジトリインターフェース
type TransferRequestRepository interface {
	// Create は新しい送金リクエストを作成
	Create(transferRequest *entities.TransferRequest) error

	// Read はIDで送金リクエストを検索
	Read(id uuid.UUID) (*entities.TransferRequest, error)

	// ReadByIdempotencyKey は冪等性キーで送金リクエストを検索
	ReadByIdempotencyKey(key string) (*entities.TransferRequest, error)

	// Update は送金リクエストを更新
	Update(transferRequest *entities.TransferRequest) error

	// ReadPendingByToUser は受取人宛の承認待ちリクエストを取得
	ReadPendingByToUser(toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error)

	// ReadSentByFromUser は送信者が送ったリクエストを取得
	ReadSentByFromUser(fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error)

	// CountPendingByToUser は受取人宛の承認待ちリクエスト数を取得
	CountPendingByToUser(toUserID uuid.UUID) (int64, error)

	// UpdateExpiredRequests は期限切れのリクエストを一括更新
	UpdateExpiredRequests() (int64, error)
}
