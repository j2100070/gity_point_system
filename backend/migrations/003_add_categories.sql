-- ================================================
-- カテゴリ管理システム
-- ================================================
-- 商品カテゴリを動的に管理するためのテーブル

-- カテゴリテーブル
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,                           -- カテゴリ名（例: 飲み物、お菓子）
    code VARCHAR(50) NOT NULL UNIQUE,                     -- カテゴリコード（例: drink, snack）
    description TEXT,                                      -- 説明
    display_order INTEGER NOT NULL DEFAULT 0,             -- 表示順序
    is_active BOOLEAN NOT NULL DEFAULT true,              -- 有効/無効
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE                   -- ソフトデリート
);

CREATE INDEX idx_categories_code ON categories(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_active ON categories(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_order ON categories(display_order) WHERE deleted_at IS NULL;

-- カテゴリのupdated_at自動更新
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ================================================
-- 初期データ（既存の商品カテゴリをマイグレーション）
-- ================================================
INSERT INTO categories (name, code, description, display_order, is_active) VALUES
('飲み物', 'drink', 'ジュースやお茶などの飲料', 1, true),
('お菓子', 'snack', 'スナックやチョコレートなどのお菓子', 2, true),
('おもちゃ', 'toy', 'ガンプラやカードゲームなどのおもちゃ', 3, true),
('その他', 'other', 'その他の商品', 99, true)
ON CONFLICT (code) DO NOTHING;

-- コメント
COMMENT ON TABLE categories IS '商品カテゴリマスタ。管理者がカテゴリを追加・編集・削除可能';
COMMENT ON COLUMN categories.code IS 'カテゴリの一意識別子。商品テーブルから参照される';
COMMENT ON COLUMN categories.display_order IS '管理画面やユーザー画面での表示順序';

-- 完了メッセージ
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'カテゴリ管理テーブルを作成しました';
    RAISE NOTICE '初期カテゴリを4件登録しました';
    RAISE NOTICE '  - 飲み物 (drink)';
    RAISE NOTICE '  - お菓子 (snack)';
    RAISE NOTICE '  - おもちゃ (toy)';
    RAISE NOTICE '  - その他 (other)';
    RAISE NOTICE '========================================';
END $$;
