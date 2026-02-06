package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// BalanceUpdate は残高更新のパラメータ
type BalanceUpdate struct {
	UserID   uuid.UUID
	Amount   int64
	IsDeduct bool // true: 減算, false: 加算
}

// UserRepository はユーザーのリポジトリインターフェース
// Interactorが要求するリポジトリの仕様を定義
type UserRepository interface {
	// Create は新しいユーザーを作成
	Create(user *entities.User) error

	// Read はIDでユーザーを検索
	Read(id uuid.UUID) (*entities.User, error)

	// ReadByUsername はユーザー名でユーザーを検索
	ReadByUsername(username string) (*entities.User, error)

	// ReadByEmail はメールアドレスでユーザーを検索
	ReadByEmail(email string) (*entities.User, error)

	// Update はユーザー情報を更新（楽観的ロック対応）
	// 返り値のboolは更新が成功したかどうか（versionが一致したか）
	Update(user *entities.User) (bool, error)

	// UpdateBalanceWithLock は残高を更新（悲観的ロック: SELECT FOR UPDATE）
	// トランザクション内で使用する
	UpdateBalanceWithLock(tx interface{}, userID uuid.UUID, amount int64, isDeduct bool) error

	// UpdateBalancesWithLock は複数ユーザーの残高を一括更新（悲観的ロック、デッドロック回避）
	// 内部でID順にロックを取得することでデッドロックを回避します
	UpdateBalancesWithLock(tx interface{}, updates []BalanceUpdate) error

	// ReadList はユーザー一覧を取得（ページネーション対応）
	ReadList(offset, limit int) ([]*entities.User, error)

	// Count はユーザー総数を取得
	Count() (int64, error)

	// Delete はユーザーを論理削除
	Delete(id uuid.UUID) error
}
