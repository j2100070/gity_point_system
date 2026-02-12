-- 007_add_daily_bonuses.sql
-- デイリーボーナステーブルを作成

CREATE TABLE IF NOT EXISTS daily_bonuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bonus_date DATE NOT NULL,

    -- 各ボーナスの達成状況
    login_completed BOOLEAN NOT NULL DEFAULT FALSE,
    login_completed_at TIMESTAMPTZ,

    transfer_completed BOOLEAN NOT NULL DEFAULT FALSE,
    transfer_completed_at TIMESTAMPTZ,
    transfer_transaction_id UUID REFERENCES transactions(id),

    exchange_completed BOOLEAN NOT NULL DEFAULT FALSE,
    exchange_completed_at TIMESTAMPTZ,
    exchange_id UUID REFERENCES product_exchanges(id),

    -- 全て達成したかどうか
    all_completed BOOLEAN NOT NULL DEFAULT FALSE,
    all_completed_at TIMESTAMPTZ,

    -- 付与されたポイント
    total_bonus_points BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- ユーザーと日付の組み合わせは一意
    UNIQUE(user_id, bonus_date)
);

-- インデックス
CREATE INDEX idx_daily_bonuses_user_date ON daily_bonuses(user_id, bonus_date DESC);
CREATE INDEX idx_daily_bonuses_date ON daily_bonuses(bonus_date DESC);
CREATE INDEX idx_daily_bonuses_user_all_completed ON daily_bonuses(user_id, all_completed);

-- コメント
COMMENT ON TABLE daily_bonuses IS 'デイリーボーナスの達成状況を記録';
COMMENT ON COLUMN daily_bonuses.bonus_date IS 'ボーナス対象日（JSTの日付）';
COMMENT ON COLUMN daily_bonuses.login_completed IS 'ログインボーナス達成フラグ';
COMMENT ON COLUMN daily_bonuses.transfer_completed IS '送金ボーナス達成フラグ';
COMMENT ON COLUMN daily_bonuses.exchange_completed IS '商品交換ボーナス達成フラグ';
COMMENT ON COLUMN daily_bonuses.all_completed IS '全ボーナス達成フラグ';
COMMENT ON COLUMN daily_bonuses.total_bonus_points IS '当日に付与された合計ボーナスポイント';
