-- ================================================
-- PayPay風 送金リクエスト機能追加
-- ================================================

-- ユーザーテーブルに個人固定QRコードを追加
ALTER TABLE users ADD COLUMN personal_qr_code VARCHAR(255);

-- 既存ユーザーに個人QRコードを生成
UPDATE users SET personal_qr_code = 'user:' || id::text WHERE personal_qr_code IS NULL;

-- NOT NULL制約を追加
ALTER TABLE users ALTER COLUMN personal_qr_code SET NOT NULL;

-- ユニーク制約を追加
ALTER TABLE users ADD CONSTRAINT users_personal_qr_code_unique UNIQUE (personal_qr_code);

-- インデックスを追加
CREATE INDEX idx_users_personal_qr_code ON users(personal_qr_code);

-- 送金リクエストテーブル
CREATE TABLE transfer_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- 送信者
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,   -- 受取人
    amount BIGINT NOT NULL CHECK (amount > 0),                         -- 送金額
    message TEXT,                                                      -- オプショナルメモ
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'expired')),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,                      -- 重複防止キー
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,                      -- 有効期限（24時間）
    approved_at TIMESTAMP WITH TIME ZONE,                              -- 承認日時
    rejected_at TIMESTAMP WITH TIME ZONE,                              -- 拒否日時
    cancelled_at TIMESTAMP WITH TIME ZONE,                             -- キャンセル日時
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL, -- 承認後のTransaction
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- ビジネスルール制約
    CONSTRAINT check_different_users_transfer_request CHECK (from_user_id != to_user_id)
);

-- インデックス
CREATE INDEX idx_transfer_requests_from_user ON transfer_requests(from_user_id);
CREATE INDEX idx_transfer_requests_to_user ON transfer_requests(to_user_id);
CREATE INDEX idx_transfer_requests_status ON transfer_requests(status);
CREATE INDEX idx_transfer_requests_expires_at ON transfer_requests(expires_at);
CREATE INDEX idx_transfer_requests_created_at ON transfer_requests(created_at DESC);
CREATE INDEX idx_transfer_requests_idempotency_key ON transfer_requests(idempotency_key);

-- 受取人の承認待ちリクエストを効率的に取得するための複合インデックス
CREATE INDEX idx_transfer_requests_to_user_status ON transfer_requests(to_user_id, status)
    WHERE status = 'pending';

-- 送信者の送信済みリクエストを効率的に取得するための複合インデックス
CREATE INDEX idx_transfer_requests_from_user_created ON transfer_requests(from_user_id, created_at DESC);

-- 更新トリガー
CREATE TRIGGER update_transfer_requests_updated_at BEFORE UPDATE ON transfer_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- コメント
COMMENT ON TABLE transfer_requests IS 'PayPay風の送金リクエスト（承認待ち送金）';
COMMENT ON COLUMN transfer_requests.expires_at IS '有効期限（作成から24時間）';
COMMENT ON COLUMN transfer_requests.status IS 'pending: 承認待ち, approved: 承認済み, rejected: 拒否, cancelled: キャンセル, expired: 期限切れ';
