package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// CategoryRepository はカテゴリのリポジトリインターフェース
type CategoryRepository interface {
	// Create は新しいカテゴリを作成
	Create(category *entities.Category) error

	// Read はIDでカテゴリを検索
	Read(id uuid.UUID) (*entities.Category, error)

	// ReadByCode はコードでカテゴリを検索
	ReadByCode(code string) (*entities.Category, error)

	// Update はカテゴリ情報を更新
	Update(category *entities.Category) error

	// Delete はカテゴリを論理削除
	Delete(id uuid.UUID) error

	// ReadList はカテゴリ一覧を取得（表示順序順）
	ReadList(activeOnly bool) ([]*entities.Category, error)

	// Count はカテゴリ総数を取得
	Count() (int64, error)

	// ExistsCode はコードが存在するか確認
	ExistsCode(code string, excludeID *uuid.UUID) (bool, error)
}
