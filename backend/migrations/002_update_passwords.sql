-- パスワードハッシュを更新
-- admin: admin123
-- testuser: test123

UPDATE users
SET password_hash = '$2a$10$Rw0nZrWoh7CvWW9HbwYM6.P251lNv94/jmhuLLhWF70Qt2vH8dthO'
WHERE username = 'admin';

UPDATE users
SET password_hash = '$2a$10$Icg8iyLTgkbpAT8TNLIxG.HigjDo4EjmqyLjELlu1XBlU0uyx8emy'
WHERE username = 'testuser';

-- 確認メッセージ
DO $$
BEGIN
    RAISE NOTICE 'Passwords updated successfully!';
    RAISE NOTICE 'admin: admin123';
    RAISE NOTICE 'testuser: test123';
END $$;
