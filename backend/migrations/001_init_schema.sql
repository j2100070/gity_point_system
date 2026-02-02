-- Point System Database Schema
-- 初期化スクリプト

-- UUID拡張を有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- トランザクションテーブル
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID REFERENCES users(id),
    to_user_id UUID REFERENCES users(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    transaction_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    idempotency_key VARCHAR(255),
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    CHECK (from_user_id IS NOT NULL OR to_user_id IS NOT NULL),
    CHECK (from_user_id IS NULL OR to_user_id IS NULL OR from_user_id != to_user_id)
);

-- 冪等性キーテーブル
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    transaction_id UUID REFERENCES transactions(id),
    status VARCHAR(20) NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'completed', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- セッションテーブル
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    csrf_token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 友達関係テーブル
CREATE TABLE IF NOT EXISTS friendships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id UUID NOT NULL REFERENCES users(id),
    addressee_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'blocked')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (requester_id != addressee_id),
    UNIQUE (requester_id, addressee_id)
);

-- QRコードテーブル
CREATE TABLE IF NOT EXISTS qr_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    code VARCHAR(255) UNIQUE NOT NULL,
    qr_type VARCHAR(20) NOT NULL CHECK (qr_type IN ('receive', 'send')),
    amount BIGINT CHECK (amount IS NULL OR amount > 0),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    used_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックス作成

-- ユーザー
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- トランザクション
CREATE INDEX IF NOT EXISTS idx_transactions_from_user ON transactions(from_user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_user ON transactions(to_user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);

-- 冪等性キー
CREATE INDEX IF NOT EXISTS idx_idempotency_user ON idempotency_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_idempotency_expires ON idempotency_keys(expires_at);

-- セッション
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- 友達関係
CREATE INDEX IF NOT EXISTS idx_friendships_requester ON friendships(requester_id);
CREATE INDEX IF NOT EXISTS idx_friendships_addressee ON friendships(addressee_id);
CREATE INDEX IF NOT EXISTS idx_friendships_status ON friendships(status);

-- QRコード
CREATE INDEX IF NOT EXISTS idx_qrcodes_user ON qr_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_qrcodes_code ON qr_codes(code);
CREATE INDEX IF NOT EXISTS idx_qrcodes_expires ON qr_codes(expires_at);

-- トリガー: updated_at自動更新
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_friendships_updated_at BEFORE UPDATE ON friendships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 初期管理ユーザー作成（パスワード: admin123）
-- bcrypt hash for "admin123" (cost=10)
INSERT INTO users (username, email, password_hash, display_name, balance, role, is_active)
VALUES (
    'admin',
    'admin@example.com',
    '$2a$10$Rw0nZrWoh7CvWW9HbwYM6.P251lNv94/jmhuLLhWF70Qt2vH8dthO',
    'System Administrator',
    1000000,
    'admin',
    true
) ON CONFLICT (username) DO NOTHING;

-- テストユーザー作成（パスワード: test123）
-- bcrypt hash for "test123" (cost=10)
INSERT INTO users (username, email, password_hash, display_name, balance, role, is_active)
VALUES (
    'testuser',
    'test@example.com',
    '$2a$10$Icg8iyLTgkbpAT8TNLIxG.HigjDo4EjmqyLjELlu1XBlU0uyx8emy',
    'Test User',
    10000,
    'user',
    true
) ON CONFLICT (username) DO NOTHING;

-- 完了メッセージ
DO $$
BEGIN
    RAISE NOTICE 'Database schema initialized successfully!';
    RAISE NOTICE 'Admin user: admin / admin123';
    RAISE NOTICE 'Test user: testuser / test123';
END $$;
