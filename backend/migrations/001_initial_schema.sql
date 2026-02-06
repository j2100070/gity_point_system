-- ================================================
-- ポイントシステム データベーススキーマ
-- ================================================
-- セキュリティ要件:
--   - SQLインジェクション: パラメータ化クエリ使用必須
--   - パスワード: bcryptでハッシュ化（cost=12以上推奨）
--   - セッション: httponly, secure, samesite=strict
--
-- トランザクション要件:
--   - 楽観的ロック: versionカラムで競合検知
--   - 冪等性: idempotency_keyで重複防止
--   - 残高整合性: SELECT FOR UPDATE使用
-- ================================================

-- ユーザーテーブル
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt hash
    display_name VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0), -- 負の残高を防止
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    version INTEGER NOT NULL DEFAULT 1, -- 楽観的ロック用
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE -- ソフトデリート
);

CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- ユーザーの更新時にupdated_atを自動更新
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- セッションテーブル (Session-based認証)
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL, -- ランダム生成トークン
    csrf_token VARCHAR(255) NOT NULL, -- CSRFトークン
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(session_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- 冪等性キーテーブル（重複トランザクション防止）
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_id UUID, -- 完了したトランザクションのID
    status VARCHAR(50) NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'completed', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '24 hours')
);

CREATE INDEX idx_idempotency_user_id ON idempotency_keys(user_id);
CREATE INDEX idx_idempotency_expires_at ON idempotency_keys(expires_at);

-- トランザクションテーブル（ポイント移動履歴）
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- 送信者（NULL=システム付与）
    to_user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- 受信者（NULL=システムへの返却）
    amount BIGINT NOT NULL CHECK (amount > 0), -- 正の値のみ
    transaction_type VARCHAR(50) NOT NULL CHECK (transaction_type IN ('transfer', 'admin_grant', 'admin_deduct', 'system_grant')),
    status VARCHAR(50) NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed', 'reversed')),
    idempotency_key VARCHAR(255) REFERENCES idempotency_keys(key),
    description TEXT,
    metadata JSONB, -- 追加情報（QRコードID、管理者メモなど）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    -- ビジネスルール制約
    CONSTRAINT check_different_users CHECK (
        transaction_type != 'transfer' OR from_user_id != to_user_id
    )
);

CREATE INDEX idx_transactions_from_user ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user ON transactions(to_user_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);

-- 友達関係テーブル
CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'blocked')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- 同じペアの重複を防止
    CONSTRAINT unique_friendship UNIQUE (requester_id, addressee_id),
    -- 自分自身との友達関係を防止
    CONSTRAINT check_different_friends CHECK (requester_id != addressee_id)
);

CREATE INDEX idx_friendships_requester ON friendships(requester_id);
CREATE INDEX idx_friendships_addressee ON friendships(addressee_id);
CREATE INDEX idx_friendships_status ON friendships(status);

CREATE TRIGGER update_friendships_updated_at BEFORE UPDATE ON friendships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- QRコードテーブル（一時的なポイント受取用）
CREATE TABLE qr_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(255) UNIQUE NOT NULL, -- ランダム生成コード
    amount BIGINT CHECK (amount > 0), -- NULL=送信者が金額指定、値あり=固定額
    qr_type VARCHAR(50) NOT NULL CHECK (qr_type IN ('receive', 'send')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    used_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_qr_codes_user_id ON qr_codes(user_id);
CREATE INDEX idx_qr_codes_code ON qr_codes(code) WHERE used_at IS NULL;
CREATE INDEX idx_qr_codes_expires_at ON qr_codes(expires_at);

-- 監査ログテーブル（管理者操作の記録）
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    target_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL, -- 'grant_points', 'deduct_points', 'change_role', etc.
    details JSONB NOT NULL, -- 操作の詳細
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_admin_user ON audit_logs(admin_user_id);
CREATE INDEX idx_audit_logs_target_user ON audit_logs(target_user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- ================================================
-- セキュリティビュー（パスワードハッシュを隠蔽）
-- ================================================
CREATE VIEW users_safe AS
SELECT
    id, username, email, display_name, balance, role,
    version, is_active, created_at, updated_at
FROM users
WHERE deleted_at IS NULL;

-- ================================================
-- 初期データ（開発用）
-- ================================================
-- 管理者ユーザー: admin / admin123
-- パスワードハッシュ: bcrypt cost=10
INSERT INTO users (username, email, password_hash, display_name, role, balance) VALUES
('admin', 'admin@example.com', '$2a$10$Rw0nZrWoh7CvWW9HbwYM6.P251lNv94/jmhuLLhWF70Qt2vH8dthO', 'System Administrator', 'admin', 1000000)
ON CONFLICT (username) DO NOTHING;

-- テストユーザー: testuser / test123
INSERT INTO users (username, email, password_hash, display_name, balance) VALUES
('testuser', 'test@example.com', '$2a$10$Icg8iyLTgkbpAT8TNLIxG.HigjDo4EjmqyLjELlu1XBlU0uyx8emy', 'Test User', 10000)
ON CONFLICT (username) DO NOTHING;

COMMENT ON TABLE users IS 'ユーザーマスタ。楽観的ロック(version)とソフトデリート対応';
COMMENT ON COLUMN users.balance IS 'ポイント残高（負の値は制約で禁止）';
COMMENT ON COLUMN users.version IS '楽観的ロック用バージョン番号。更新時に必ずインクリメント';
COMMENT ON TABLE sessions IS 'セッション管理。Session-based認証とCSRF対策';
COMMENT ON TABLE idempotency_keys IS '冪等性保証。同一キーでの重複トランザクションを防止';
COMMENT ON TABLE transactions IS 'ポイント移動履歴。全ての残高変更を記録';
COMMENT ON TABLE friendships IS '友達関係。双方向の関係性を管理';
COMMENT ON TABLE qr_codes IS 'QRコード管理。一時的な受取・送信用';
COMMENT ON TABLE audit_logs IS '監査ログ。管理者操作の完全な記録';

-- ================================================
-- 完了メッセージ
-- ================================================
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Database schema initialized successfully!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Admin user: admin / admin123';
    RAISE NOTICE 'Test user: testuser / test123';
    RAISE NOTICE '========================================';
END $$;
