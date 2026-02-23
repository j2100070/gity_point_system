-- ================================================
-- バルクシードスクリプト: 100万トランザクション生成
-- ================================================
-- 前提:
--   - seed_100k_users.sql が実行済みで、user_1 〜 user_100000 が存在する
--   - admin ユーザーが存在する
--
-- 使い方:
--   docker exec -i gity_point_system-db-1 psql -U postgres -d gity_point < backend/scripts/seed_1m_transactions.sql
--
-- 注意:
--   - 実行時間: 約2〜5分（idempotency_keys + transactions）
--   - 残高は更新しない（トランザクション履歴のみ生成）
-- ================================================

DO $$
DECLARE
    user_ids UUID[];
    admin_id UUID;
    total_users INTEGER;
    inserted_count INTEGER;
    start_time TIMESTAMP;
BEGIN
    start_time := clock_timestamp();
    RAISE NOTICE 'Starting bulk transaction seed...';

    -- ユーザーIDを配列に取得（ランダムアクセス用）
    SELECT array_agg(id ORDER BY username)
    INTO user_ids
    FROM (SELECT id, username FROM users WHERE role = 'user' LIMIT 10000) sub;

    -- adminユーザーIDを取得
    SELECT id INTO admin_id FROM users WHERE role = 'admin' LIMIT 1;

    total_users := array_length(user_ids, 1);
    RAISE NOTICE 'Found % users for seeding', total_users;

    IF total_users IS NULL OR total_users < 2 THEN
        RAISE EXCEPTION 'Not enough users. Run seed_100k_users.sql first.';
    END IF;

    -- ========================================
    -- 1. transfer（40万件）: ユーザー間送金
    -- ========================================
    RAISE NOTICE 'Inserting 400,000 transfer idempotency_keys...';

    INSERT INTO idempotency_keys (key, user_id, status, expires_at)
    SELECT
        'seed-transfer-' || i,
        user_ids[1 + (i % total_users)],
        'completed',
        NOW() + INTERVAL '24 hours'
    FROM generate_series(1, 400000) AS i
    ON CONFLICT (key) DO NOTHING;

    RAISE NOTICE 'Inserting 400,000 transfer transactions...';

    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, status, idempotency_key, description, created_at, completed_at)
    SELECT
        user_ids[1 + (i % total_users)],
        user_ids[1 + ((i + 1 + (i % 97)) % total_users)],  -- 異なるユーザーを保証
        (50 + (i % 951))::BIGINT,    -- 50〜1000ポイント
        'transfer',
        'completed',
        'seed-transfer-' || i,
        'Seed transfer #' || i,
        NOW() - (i || ' minutes')::INTERVAL,
        NOW() - (i || ' minutes')::INTERVAL
    FROM generate_series(1, 400000) AS i;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RAISE NOTICE '  Inserted: % transfer transactions', inserted_count;

    -- ========================================
    -- 2. daily_bonus（30万件）: デイリーボーナス
    -- ========================================
    RAISE NOTICE 'Inserting 300,000 daily_bonus idempotency_keys...';

    INSERT INTO idempotency_keys (key, user_id, status, expires_at)
    SELECT
        'seed-daily-' || i,
        user_ids[1 + (i % total_users)],
        'completed',
        NOW() + INTERVAL '24 hours'
    FROM generate_series(1, 300000) AS i
    ON CONFLICT (key) DO NOTHING;

    RAISE NOTICE 'Inserting 300,000 daily_bonus transactions...';

    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, status, idempotency_key, description, created_at, completed_at)
    SELECT
        NULL,  -- システムから付与
        user_ids[1 + (i % total_users)],
        CASE
            WHEN i % 100 < 5  THEN 100   -- 5%: 大当たり
            WHEN i % 100 < 20 THEN 50    -- 15%: 当たり
            WHEN i % 100 < 50 THEN 10    -- 30%: 小当たり
            ELSE 5                        -- 50%: 通常
        END,
        'daily_bonus',
        'completed',
        'seed-daily-' || i,
        'Daily bonus',
        NOW() - (i || ' minutes')::INTERVAL,
        NOW() - (i || ' minutes')::INTERVAL
    FROM generate_series(1, 300000) AS i;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RAISE NOTICE '  Inserted: % daily_bonus transactions', inserted_count;

    -- ========================================
    -- 3. admin_grant（15万件）: 管理者ポイント付与
    -- ========================================
    RAISE NOTICE 'Inserting 150,000 admin_grant idempotency_keys...';

    INSERT INTO idempotency_keys (key, user_id, status, expires_at)
    SELECT
        'seed-agrant-' || i,
        admin_id,
        'completed',
        NOW() + INTERVAL '24 hours'
    FROM generate_series(1, 150000) AS i
    ON CONFLICT (key) DO NOTHING;

    RAISE NOTICE 'Inserting 150,000 admin_grant transactions...';

    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, status, idempotency_key, description, metadata, created_at, completed_at)
    SELECT
        admin_id,
        user_ids[1 + (i % total_users)],
        (100 + (i % 901))::BIGINT,  -- 100〜1000ポイント
        'admin_grant',
        'completed',
        'seed-agrant-' || i,
        'Admin grant #' || i,
        '{"reason": "seed data"}'::JSONB,
        NOW() - (i || ' minutes')::INTERVAL,
        NOW() - (i || ' minutes')::INTERVAL
    FROM generate_series(1, 150000) AS i;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RAISE NOTICE '  Inserted: % admin_grant transactions', inserted_count;

    -- ========================================
    -- 4. admin_deduct（5万件）: 管理者ポイント減算
    -- ========================================
    RAISE NOTICE 'Inserting 50,000 admin_deduct idempotency_keys...';

    INSERT INTO idempotency_keys (key, user_id, status, expires_at)
    SELECT
        'seed-adeduct-' || i,
        admin_id,
        'completed',
        NOW() + INTERVAL '24 hours'
    FROM generate_series(1, 50000) AS i
    ON CONFLICT (key) DO NOTHING;

    RAISE NOTICE 'Inserting 50,000 admin_deduct transactions...';

    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, status, idempotency_key, description, metadata, created_at, completed_at)
    SELECT
        user_ids[1 + (i % total_users)],
        NULL,  -- システムへ返却
        (10 + (i % 491))::BIGINT,  -- 10〜500ポイント
        'admin_deduct',
        'completed',
        'seed-adeduct-' || i,
        'Admin deduct #' || i,
        '{"reason": "seed data"}'::JSONB,
        NOW() - (i || ' minutes')::INTERVAL,
        NOW() - (i || ' minutes')::INTERVAL
    FROM generate_series(1, 50000) AS i;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RAISE NOTICE '  Inserted: % admin_deduct transactions', inserted_count;

    -- ========================================
    -- 5. system_grant（10万件）: システムポイント付与
    -- ========================================
    RAISE NOTICE 'Inserting 100,000 system_grant idempotency_keys...';

    INSERT INTO idempotency_keys (key, user_id, status, expires_at)
    SELECT
        'seed-sgrant-' || i,
        user_ids[1 + (i % total_users)],
        'completed',
        NOW() + INTERVAL '24 hours'
    FROM generate_series(1, 100000) AS i
    ON CONFLICT (key) DO NOTHING;

    RAISE NOTICE 'Inserting 100,000 system_grant transactions...';

    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, status, idempotency_key, description, created_at, completed_at)
    SELECT
        NULL,
        user_ids[1 + (i % total_users)],
        (10 + (i % 91))::BIGINT,  -- 10〜100ポイント
        'system_grant',
        'completed',
        'seed-sgrant-' || i,
        'System grant #' || i,
        NOW() - (i || ' minutes')::INTERVAL,
        NOW() - (i || ' minutes')::INTERVAL
    FROM generate_series(1, 100000) AS i;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RAISE NOTICE '  Inserted: % system_grant transactions', inserted_count;

    -- ========================================
    -- idempotency_keys に transaction_id を紐付け
    -- ========================================
    RAISE NOTICE 'Linking transaction_ids to idempotency_keys...';

    UPDATE idempotency_keys ik
    SET transaction_id = t.id
    FROM transactions t
    WHERE t.idempotency_key = ik.key
      AND ik.key LIKE 'seed-%'
      AND ik.transaction_id IS NULL;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RAISE NOTICE '  Updated: % idempotency_keys with transaction_id', inserted_count;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'Bulk transaction seed completed!';
    RAISE NOTICE '  Duration: %', clock_timestamp() - start_time;
    RAISE NOTICE '========================================';
END $$;

-- 確認クエリ
SELECT COUNT(*) AS total_transactions FROM transactions;
SELECT transaction_type, COUNT(*) AS count FROM transactions GROUP BY transaction_type ORDER BY count DESC;
SELECT COUNT(*) AS total_idempotency_keys FROM idempotency_keys WHERE key LIKE 'seed-%';
