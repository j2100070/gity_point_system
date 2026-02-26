-- ================================================
-- バルクシードスクリプト: 100万フレンドシップ生成
-- ================================================
-- 前提:
--   - seed_100k_users.sql が実行済みで、user_1 〜 user_100000 が存在する
--
-- 使い方:
--   docker exec -i gity_point_system-db-1 psql -U postgres -d gity_point < backend/scripts/seed_1m_friendships.sql
--
-- 注意:
--   - 実行時間: 約1〜3分
--   - UNIQUE(requester_id, addressee_id) 制約あり
--   - requester_id != addressee_id 制約あり
-- ================================================

DO $$
DECLARE
    user_ids UUID[];
    total_users INTEGER;
    inserted_count INTEGER;
    batch_inserted INTEGER;
    start_time TIMESTAMP;
    batch_start INTEGER;
    batch_size INTEGER := 100000;
    target_total INTEGER := 1000000;
BEGIN
    start_time := clock_timestamp();
    RAISE NOTICE 'Starting bulk friendship seed...';

    -- ユーザーIDを配列に取得
    SELECT array_agg(id ORDER BY username)
    INTO user_ids
    FROM (SELECT id, username FROM users WHERE role = 'user' LIMIT 10000) sub;

    total_users := array_length(user_ids, 1);
    RAISE NOTICE 'Found % users for seeding', total_users;

    IF total_users IS NULL OR total_users < 100 THEN
        RAISE EXCEPTION 'Not enough users (need at least 100). Run seed_100k_users.sql first.';
    END IF;

    inserted_count := 0;

    -- 10バッチに分けてINSERT（メモリ効率化）
    FOR batch_num IN 0..9 LOOP
        batch_start := batch_num * batch_size;

        RAISE NOTICE 'Inserting batch % (% - %)...',
            batch_num + 1, batch_start + 1, batch_start + batch_size;

        INSERT INTO friendships (requester_id, addressee_id, status, created_at, updated_at)
        SELECT
            user_ids[1 + (i % total_users)],
            user_ids[1 + (((i / total_users) + 1 + (i % (total_users - 1))) % total_users)],
            CASE
                WHEN i % 100 < 70 THEN 'accepted'   -- 70%: 承認済み
                WHEN i % 100 < 90 THEN 'pending'     -- 20%: 保留中
                WHEN i % 100 < 97 THEN 'rejected'    -- 7%:  拒否
                ELSE 'blocked'                        -- 3%:  ブロック
            END,
            NOW() - ((i % 365) || ' days')::INTERVAL - ((i % 24) || ' hours')::INTERVAL,
            NOW() - ((i % 180) || ' days')::INTERVAL
        FROM generate_series(batch_start + 1, batch_start + batch_size) AS i
        ON CONFLICT (requester_id, addressee_id) DO NOTHING;

        GET DIAGNOSTICS batch_inserted = ROW_COUNT;
        inserted_count := inserted_count + batch_inserted;
        RAISE NOTICE '  Batch %: inserted % (total so far: %)',
            batch_num + 1, batch_inserted, inserted_count;
    END LOOP;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'Bulk friendship seed completed!';
    RAISE NOTICE '  Inserted: % friendships', inserted_count;
    RAISE NOTICE '  Duration: %', clock_timestamp() - start_time;
    RAISE NOTICE '========================================';
END $$;

-- 確認クエリ
SELECT COUNT(*) AS total_friendships FROM friendships;
SELECT status, COUNT(*) AS count FROM friendships GROUP BY status ORDER BY count DESC;
