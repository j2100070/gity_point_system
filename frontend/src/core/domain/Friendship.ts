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
  };
  addressee?: {
    id: string;
    username: string;
    display_name: string;
  };
}

export interface FriendRequestRequest {
  addressee_id: string;
}

export interface FriendActionRequest {
  friendship_id: string;
}
