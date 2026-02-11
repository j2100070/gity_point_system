// フラットなFriendship（フレンドリスト用）
export interface FriendInfo {
  friendship: {
    id: string;
    requester_id: string;
    addressee_id: string;
    status: 'pending' | 'accepted' | 'rejected' | 'blocked';
    created_at: string;
    updated_at: string;
  };
  friend: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
  };
}

// 保留中申請（リクエスト用）
export interface PendingRequestInfo {
  friendship: {
    id: string;
    requester_id: string;
    addressee_id: string;
    status: 'pending' | 'accepted' | 'rejected' | 'blocked';
    created_at: string;
    updated_at: string;
  };
  requester: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
  };
}

// 後方互換のためFriendship型も維持
export interface Friendship {
  id: string;
  requester_id: string;
  addressee_id: string;
  status: 'pending' | 'accepted' | 'rejected' | 'blocked';
  created_at: string;
  updated_at: string;
  requester?: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
  };
  addressee?: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
  };
}

export interface FriendRequestRequest {
  addressee_id: string;
}

export interface FriendActionRequest {
  friendship_id: string;
}
