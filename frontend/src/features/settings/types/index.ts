// ユーザープロフィール型
export interface UserProfile {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  email_verified: boolean;
  email_verified_at?: string;
  balance: number;
  role: string;
  created_at: string;
}

// プロフィール更新リクエスト
export interface UpdateProfileRequest {
  display_name: string;
  email: string;
}

// プロフィール更新レスポンス
export interface UpdateProfileResponse {
  message: string;
  user: UserProfile;
  email_verification_sent?: boolean;
}

// ユーザー名変更リクエスト
export interface UpdateUsernameRequest {
  new_username: string;
}

// パスワード変更リクエスト
export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

// アバターアップロードレスポンス
export interface UploadAvatarResponse {
  message: string;
  avatar_url: string;
}

// メール認証リクエスト
export interface VerifyEmailRequest {
  token: string;
}

// メール認証レスポンス
export interface VerifyEmailResponse {
  message: string;
  user: UserProfile;
}

// アカウント削除リクエスト
export interface ArchiveAccountRequest {
  password: string;
  deletion_reason?: string;
}

// 共通レスポンス型
export interface ApiResponse {
  message: string;
}

// プロフィール取得レスポンス
export interface GetProfileResponse {
  user: UserProfile;
}
