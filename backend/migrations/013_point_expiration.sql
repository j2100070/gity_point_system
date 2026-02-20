-- 013_point_expiration.sql
-- ポイント有効期限（3ヶ月）の実装

-- ポイントバッチテーブル: 獲得ポイントをバッチ単位で追跡し、FIFO消費と期限切れ処理を行う
CREATE TABLE IF NOT EXISTS point_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_amount BIGINT NOT NULL CHECK (original_amount > 0),
    remaining_amount BIGINT NOT NULL CHECK (remaining_amount >= 0),
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('transfer', 'admin_grant', 'daily_bonus', 'system_grant', 'migration')),
    source_transaction_id UUID REFERENCES transactions(id),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- FIFO消費クエリ用: ユーザーの有効なバッチを古い順に取得
CREATE INDEX IF NOT EXISTS idx_point_batches_fifo
    ON point_batches(user_id, created_at ASC)
    WHERE remaining_amount > 0;

-- 失効ワーカー用: 期限切れで残量があるバッチを検索
CREATE INDEX IF NOT EXISTS idx_point_batches_expiry
    ON point_batches(expires_at)
    WHERE remaining_amount > 0;

-- transaction_typeにsystem_expireを追加
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_transaction_type_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_transaction_type_check
    CHECK (transaction_type IN ('transfer', 'admin_grant', 'admin_deduct', 'system_grant', 'daily_bonus', 'system_expire'));

-- 既存ユーザーの残高を初期バッチとして作成（balance > 0のユーザーのみ）
INSERT INTO point_batches (user_id, original_amount, remaining_amount, source_type, expires_at)
SELECT id, balance, balance, 'migration', NOW() + INTERVAL '3 months'
FROM users
WHERE balance > 0
ON CONFLICT DO NOTHING;

COMMENT ON TABLE point_batches IS 'ポイントバッチ: 獲得ポイントの有効期限追跡。FIFO消費と3ヶ月期限切れ処理に使用';
