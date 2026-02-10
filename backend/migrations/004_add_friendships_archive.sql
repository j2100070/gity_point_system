-- ================================================
-- フレンド関係アーカイブテーブル
-- フレンド解散時にレコードを保存
-- ================================================

CREATE TABLE IF NOT EXISTS friendships_archive (
    id UUID PRIMARY KEY,
    requester_id UUID NOT NULL,
    addressee_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    archived_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_by UUID  -- 解散を実行したユーザーID
);

CREATE INDEX idx_friendships_archive_requester ON friendships_archive(requester_id);
CREATE INDEX idx_friendships_archive_addressee ON friendships_archive(addressee_id);
CREATE INDEX idx_friendships_archive_archived_at ON friendships_archive(archived_at DESC);

COMMENT ON TABLE friendships_archive IS 'フレンド関係アーカイブ。解散された友達関係の履歴を保持';
