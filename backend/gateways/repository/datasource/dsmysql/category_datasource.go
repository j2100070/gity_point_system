package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// CategoryDataSource はカテゴリのデータソースインターフェース
type CategoryDataSource interface {
	// Insert は新しいカテゴリを挿入
	Insert(category *entities.Category) error

	// Select はIDでカテゴリを検索
	Select(id uuid.UUID) (*entities.Category, error)

	// SelectByCode はコードでカテゴリを検索
	SelectByCode(code string) (*entities.Category, error)

	// Update はカテゴリ情報を更新
	Update(category *entities.Category) error

	// Delete はカテゴリを論理削除
	Delete(id uuid.UUID) error

	// SelectList はカテゴリ一覧を取得
	SelectList(activeOnly bool) ([]*entities.Category, error)

	// Count はカテゴリ総数を取得
	Count() (int64, error)

	// ExistsCode はコードの存在確認
	ExistsCode(code string, excludeID *uuid.UUID) (bool, error)
}
