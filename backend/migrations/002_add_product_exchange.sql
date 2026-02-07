-- ================================================
-- 商品交換システム
-- ================================================
-- お菓子・ジュースなどの商品とポイント交換機能を追加

-- 商品テーブル
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,                          -- 商品名（例: コカ・コーラ500ml）
    description TEXT,                                     -- 商品説明
    category VARCHAR(100) NOT NULL,                       -- カテゴリ（snack, drink, other）
    price BIGINT NOT NULL CHECK (price > 0),             -- 交換に必要なポイント数
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0), -- 在庫数（-1 = 無制限）
    image_url TEXT,                                       -- 商品画像URL
    is_available BOOLEAN NOT NULL DEFAULT true,           -- 交換可能フラグ
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE                   -- ソフトデリート
);

CREATE INDEX idx_products_category ON products(category) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_available ON products(is_available) WHERE deleted_at IS NULL;

-- 商品交換履歴テーブル
CREATE TABLE product_exchanges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL CHECK (quantity > 0),       -- 交換数量
    points_used BIGINT NOT NULL CHECK (points_used > 0),  -- 使用したポイント数
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled', 'delivered')),
    transaction_id UUID REFERENCES transactions(id),      -- 関連するトランザクション
    notes TEXT,                                           -- 備考（受取場所、時間など）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_product_exchanges_user ON product_exchanges(user_id);
CREATE INDEX idx_product_exchanges_product ON product_exchanges(product_id);
CREATE INDEX idx_product_exchanges_status ON product_exchanges(status);
CREATE INDEX idx_product_exchanges_created_at ON product_exchanges(created_at DESC);

-- 商品のupdated_at自動更新
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- コメント
COMMENT ON TABLE products IS '商品マスタ。お菓子・ジュースなどの交換可能商品';
COMMENT ON COLUMN products.price IS 'ポイント交換に必要なポイント数';
COMMENT ON COLUMN products.stock IS '在庫数。-1は無制限を表す';
COMMENT ON TABLE product_exchanges IS '商品交換履歴。ポイントを使った商品購入記録';

-- ================================================
-- 初期データ（サンプル商品）
-- ================================================
INSERT INTO products (name, description, category, price, stock, is_available) VALUES
-- 飲み物
('コカ・コーラ 500ml', '定番の炭酸飲料', 'drink', 100, 50, true),
('ポカリスエット 500ml', 'スポーツドリンク', 'drink', 120, 30, true),
('お〜いお茶 500ml', '緑茶飲料', 'drink', 90, 40, true),
('カルピス 500ml', '乳酸菌飲料', 'drink', 110, 25, true),
-- お菓子
('ポテトチップス うすしお', 'カルビーのポテチ', 'snack', 150, 100, true),
('じゃがりこ サラダ', 'スナック菓子', 'snack', 140, 80, true),
('キットカット', 'チョコレート菓子', 'snack', 120, 60, true),
('ブラックサンダー', 'チョコバー', 'snack', 80, 120, true),
('うまい棒 めんたい味', '駄菓子', 'snack', 20, 200, true),
('ハーゲンダッツ バニラ', 'プレミアムアイス', 'snack', 300, 20, true),
-- おもちゃ
('ガンプラ HG', 'ガンダムプラモデル', 'toy', 800, 15, true),
('トミカ ミニカー', 'ダイキャストカー', 'toy', 400, 30, true),
('ポケモンカード パック', '1パック5枚入り', 'toy', 180, 50, true),
('遊戯王カード パック', '1パック5枚入り', 'toy', 180, 50, true),
('レゴブロック 基本セット', '創造力を育むブロック', 'toy', 1200, 10, true),
('ルービックキューブ', '3x3キューブ', 'toy', 500, 20, true);

-- 完了メッセージ
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '商品交換システムのテーブルを作成しました';
    RAISE NOTICE 'サンプル商品を16件登録しました';
    RAISE NOTICE '  - 飲み物: 4件';
    RAISE NOTICE '  - お菓子: 6件';
    RAISE NOTICE '  - おもちゃ: 6件';
    RAISE NOTICE '========================================';
END $$;
