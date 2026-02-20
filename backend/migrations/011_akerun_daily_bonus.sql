-- 011_akerun_daily_bonus.sql
-- デイリーボーナスをAkerun入退室連携に刷新

-- 既存テーブルをリネーム（データ保全）: 旧スキーマ（all_completedカラムがある場合）のみ
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'daily_bonuses' AND column_name = 'all_completed'
    ) THEN
        ALTER TABLE daily_bonuses RENAME TO daily_bonuses_legacy;
    END IF;
END $$;

-- 新デイリーボーナステーブル（Akerun入退室ベース）
CREATE TABLE IF NOT EXISTS daily_bonuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bonus_date DATE NOT NULL,
    bonus_points BIGINT NOT NULL DEFAULT 5,
    akerun_access_id TEXT,
    akerun_user_name TEXT,
    accessed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, bonus_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_bonuses_user_date ON daily_bonuses(user_id, bonus_date DESC);
CREATE INDEX IF NOT EXISTS idx_daily_bonuses_date ON daily_bonuses(bonus_date DESC);

COMMENT ON TABLE daily_bonuses IS 'Akerun入退室ベースのデイリーボーナス記録';
COMMENT ON COLUMN daily_bonuses.bonus_date IS 'ボーナス対象日（JST AM6:00区切り）';
COMMENT ON COLUMN daily_bonuses.akerun_access_id IS 'Akerun履歴ID';
COMMENT ON COLUMN daily_bonuses.akerun_user_name IS 'マッチしたAkerunユーザー名';

-- AkerunポーリングState管理（シングルトン行）
CREATE TABLE IF NOT EXISTS akerun_poll_state (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_polled_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 初期行を挿入
INSERT INTO akerun_poll_state (id, last_polled_at) VALUES (1, NOW()) ON CONFLICT DO NOTHING;

COMMENT ON TABLE akerun_poll_state IS 'Akerun APIポーリングの状態管理（シングルトン）';

-- システム設定テーブル（入退室ボーナスポイント数などを管理者が設定）
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- デフォルト設定を挿入
INSERT INTO system_settings (key, value, description) VALUES
    ('akerun_bonus_points', '5', 'Akerun入退室ボーナスのポイント数')
ON CONFLICT DO NOTHING;

COMMENT ON TABLE system_settings IS 'システム設定（管理者が変更可能）';
