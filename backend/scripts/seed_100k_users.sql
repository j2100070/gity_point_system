-- ================================================
-- バルクシードスクリプト: 10万人ユーザー生成
-- ================================================
-- 使い方:
--   docker exec -i gity_point_system-db-1 psql -U postgres -d gity_point < backend/scripts/seed_100k_users.sql
--
-- 注意:
--   - 既にusernameが存在する場合はスキップ（ON CONFLICT）
--   - パスワードは全員 test123（bcrypt cost=10）
--   - 実行時間: 約10〜30秒
-- ================================================

DO $$
DECLARE
    inserted_count INTEGER;
    start_time TIMESTAMP;
BEGIN
    start_time := clock_timestamp();
    RAISE NOTICE 'Starting bulk user seed...';

    INSERT INTO users (
        username,
        email,
        password_hash,
        display_name,
        first_name,
        last_name,
        balance,
        role,
        is_active,
        personal_qr_code
    )
    SELECT
        'user_' || i,
        'user_' || i || '@example.com',
        -- test123 のbcryptハッシュ (cost=10)
        '$2a$10$Icg8iyLTgkbpAT8TNLIxG.HigjDo4EjmqyLjELlu1XBlU0uyx8emy',
        'ユーザー' || i,
        CASE
            WHEN i % 5 = 0 THEN '太郎'
            WHEN i % 5 = 1 THEN '花子'
            WHEN i % 5 = 2 THEN '一郎'
            WHEN i % 5 = 3 THEN '美咲'
            ELSE '健太'
        END,
        CASE
            WHEN i % 7 = 0 THEN '田中'
            WHEN i % 7 = 1 THEN '佐藤'
            WHEN i % 7 = 2 THEN '鈴木'
            WHEN i % 7 = 3 THEN '高橋'
            WHEN i % 7 = 4 THEN '伊藤'
            WHEN i % 7 = 5 THEN '渡辺'
            ELSE '山本'
        END,
        1000,           -- 初期ポイント
        'user',         -- ロール
        true,           -- アクティブ
        'user:seed_' || gen_random_uuid()  -- ユニークQRコード
    FROM generate_series(1, 100000) AS i
    ON CONFLICT (username) DO NOTHING;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'Bulk seed completed!';
    RAISE NOTICE '  Inserted: % users', inserted_count;
    RAISE NOTICE '  Duration: %', clock_timestamp() - start_time;
    RAISE NOTICE '========================================';
END $$;

-- 確認クエリ
SELECT COUNT(*) AS total_users FROM users;
SELECT role, COUNT(*) AS count FROM users GROUP BY role;
