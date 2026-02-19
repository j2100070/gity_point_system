-- 012_lottery_bonus.sql
-- Akerun入退室ボーナスを抽選（くじ引き）方式に変更

-- 抽選ティアテーブル
CREATE TABLE IF NOT EXISTS bonus_lottery_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    points BIGINT NOT NULL DEFAULT 0,
    probability DECIMAL(5,2) NOT NULL DEFAULT 0,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bonus_lottery_tiers_active ON bonus_lottery_tiers(is_active, display_order);

COMMENT ON TABLE bonus_lottery_tiers IS 'ボーナス抽選ティア（確率テーブル）';
COMMENT ON COLUMN bonus_lottery_tiers.name IS 'ティア名（大当たり、当たり、通常など）';
COMMENT ON COLUMN bonus_lottery_tiers.points IS '獲得ポイント（0=ハズレ）';
COMMENT ON COLUMN bonus_lottery_tiers.probability IS '当選確率（%）例: 1.00 = 1%';
COMMENT ON COLUMN bonus_lottery_tiers.display_order IS '表示順';

-- daily_bonuses に抽選結果カラム追加
ALTER TABLE daily_bonuses ADD COLUMN IF NOT EXISTS lottery_tier_id UUID REFERENCES bonus_lottery_tiers(id);
ALTER TABLE daily_bonuses ADD COLUMN IF NOT EXISTS lottery_tier_name VARCHAR(50);
ALTER TABLE daily_bonuses ADD COLUMN IF NOT EXISTS is_viewed BOOLEAN NOT NULL DEFAULT false;

-- デフォルトティア挿入
INSERT INTO bonus_lottery_tiers (name, points, probability, display_order) VALUES
    ('大当たり', 100, 1.00, 1),
    ('当たり', 10, 10.00, 2),
    ('通常', 5, 89.00, 3)
ON CONFLICT DO NOTHING;
