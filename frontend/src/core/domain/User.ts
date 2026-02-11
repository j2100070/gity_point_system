export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  balance: number;
  role: 'user' | 'admin';
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  display_name: string;
}

export interface AuthResponse {
  message: string;
  user: User;
  csrf_token: string;
}
