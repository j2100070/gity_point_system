import { axiosInstance } from '@/infrastructure/api/client';
import {
  UserProfile,
  UpdateProfileRequest,
  UpdateProfileResponse,
  UpdateUsernameRequest,
  ChangePasswordRequest,
  UploadAvatarResponse,
  VerifyEmailRequest,
  VerifyEmailResponse,
  ArchiveAccountRequest,
  ApiResponse,
  GetProfileResponse,
} from '../types';

export class SettingsRepository {
  /**
   * プロフィール情報を取得
   */
  async getProfile(): Promise<UserProfile> {
    const response = await axiosInstance.get<GetProfileResponse>('/api/settings/profile');
    return response.data.user;
  }

  /**
   * プロフィールを更新（表示名、メールアドレス）
   */
  async updateProfile(request: UpdateProfileRequest): Promise<UpdateProfileResponse> {
    const response = await axiosInstance.put<UpdateProfileResponse>('/api/settings/profile', request);
    return response.data;
  }

  /**
   * ユーザー名を変更
   */
  async updateUsername(request: UpdateUsernameRequest): Promise<ApiResponse> {
    const response = await axiosInstance.put<ApiResponse>('/api/settings/username', request);
    return response.data;
  }

  /**
   * パスワードを変更
   */
  async changePassword(request: ChangePasswordRequest): Promise<ApiResponse> {
    const response = await axiosInstance.put<ApiResponse>('/api/settings/password', request);
    return response.data;
  }

  /**
   * アバター画像をアップロード
   */
  async uploadAvatar(file: File): Promise<UploadAvatarResponse> {
    const formData = new FormData();
    formData.append('avatar', file);

    // axiosのデフォルトContent-Type: application/jsonがFormDataのboundary生成を妨げるため、
    // native fetch APIを使用してmultipart/form-dataを正しく送信する
    const csrfToken = localStorage.getItem('csrf_token') || '';
    const response = await fetch('/api/settings/avatar', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'X-CSRF-Token': csrfToken,
      },
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'upload failed' }));
      throw { response: { data: errorData, status: response.status } };
    }

    return response.json();
  }

  /**
   * アバターを削除（デフォルトに戻す）
   */
  async deleteAvatar(): Promise<ApiResponse> {
    const response = await axiosInstance.delete<ApiResponse>('/api/settings/avatar');
    return response.data;
  }

  /**
   * メール認証メールを送信
   */
  async sendEmailVerification(): Promise<ApiResponse> {
    const response = await axiosInstance.post<ApiResponse>('/api/settings/email/verify');
    return response.data;
  }

  /**
   * メールアドレスを認証
   */
  async verifyEmail(request: VerifyEmailRequest): Promise<VerifyEmailResponse> {
    const response = await axiosInstance.post<VerifyEmailResponse>('/api/settings/email/verify/confirm', request);
    return response.data;
  }

  /**
   * アカウントを削除（アーカイブ）
   */
  async archiveAccount(request: ArchiveAccountRequest): Promise<ApiResponse> {
    const response = await axiosInstance.delete<ApiResponse>('/api/settings/account', {
      data: request,
    });
    return response.data;
  }
}
