import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { SettingsRepository } from '../api/SettingsRepository';
import { UserProfile } from '../types';
import { useAuthStore } from '@/shared/stores/authStore';

const settingsRepo = new SettingsRepository();

export const SettingsPage: React.FC = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { loadUser } = useAuthStore();

  // タブ状態
  const [activeTab, setActiveTab] = useState<'profile' | 'security' | 'account'>('profile');

  // プロフィール編集
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileSuccess, setProfileSuccess] = useState('');

  // ユーザー名変更
  const [newUsername, setNewUsername] = useState('');
  const [usernameLoading, setUsernameLoading] = useState(false);
  const [usernameSuccess, setUsernameSuccess] = useState('');

  // パスワード変更
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState('');

  // アバターアップロード
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [avatarLoading, setAvatarLoading] = useState(false);

  // アカウント削除
  const [deletePassword, setDeletePassword] = useState('');
  const [deleteReason, setDeleteReason] = useState('');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      const data = await settingsRepo.getProfile();
      setProfile(data);
      setDisplayName(data.display_name);
      setEmail(data.email);
      setFirstName(data.first_name || '');
      setLastName(data.last_name || '');
      setLoading(false);
    } catch (err: any) {
      setError(err.response?.data?.error || 'プロフィールの読み込みに失敗しました');
      setLoading(false);
    }
  };

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setProfileSuccess('');
    setProfileLoading(true);

    console.log('[SettingsPage] Updating profile with current state:', {
      profile_version: (profile as any)?.version,
      display_name: displayName,
      email: email,
    });

    try {
      const response = await settingsRepo.updateProfile({
        display_name: displayName,
        email: email,
        first_name: firstName,
        last_name: lastName,
      });
      console.log('[SettingsPage] Profile updated successfully:', {
        new_version: (response.user as any).version,
      });
      setProfile(response.user);
      // フォームの値も更新されたプロフィールで同期
      setDisplayName(response.user.display_name);
      setEmail(response.user.email);
      setFirstName(response.user.first_name || '');
      setLastName(response.user.last_name || '');
      setProfileSuccess(response.message);
      if (response.email_verification_sent) {
        setProfileSuccess(response.message + ' メール認証が必要です。');
      }
    } catch (err: any) {
      console.error('[SettingsPage] Profile update failed:', err);
      setError(err.response?.data?.error || 'プロフィールの更新に失敗しました');
    } finally {
      setProfileLoading(false);
    }
  };

  const handleUpdateUsername = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setUsernameSuccess('');
    setUsernameLoading(true);

    try {
      const response = await settingsRepo.updateUsername({
        new_username: newUsername,
      });
      setUsernameSuccess(response.message);
      setNewUsername('');
      await loadProfile();
    } catch (err: any) {
      setError(err.response?.data?.error || 'ユーザー名の変更に失敗しました');
    } finally {
      setUsernameLoading(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setPasswordSuccess('');

    if (newPassword !== confirmPassword) {
      setError('新しいパスワードが一致しません');
      return;
    }

    if (newPassword.length < 8) {
      setError('パスワードは8文字以上にしてください');
      return;
    }

    setPasswordLoading(true);
    try {
      const response = await settingsRepo.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setPasswordSuccess(response.message);
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'パスワードの変更に失敗しました');
    } finally {
      setPasswordLoading(false);
    }
  };

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setAvatarFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setAvatarPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleUploadAvatar = async () => {
    if (!avatarFile) return;

    setError('');
    setAvatarLoading(true);
    try {
      const response = await settingsRepo.uploadAvatar(avatarFile);
      console.log('[SettingsPage] Avatar uploaded successfully');
      setProfileSuccess(response.message);
      const updatedProfile = await settingsRepo.getProfile();
      console.log('[SettingsPage] Profile reloaded after avatar upload:', {
        version: (updatedProfile as any).version,
        display_name: updatedProfile.display_name,
        email: updatedProfile.email,
      });
      setProfile(updatedProfile);
      // フォームの値も更新されたプロフィールで同期
      setDisplayName(updatedProfile.display_name);
      setEmail(updatedProfile.email);
      setAvatarFile(null);
      setAvatarPreview(null);
      // authStoreのユーザー情報も更新（ヘッダーのアバター表示用）
      await loadUser();
    } catch (err: any) {
      console.error('[SettingsPage] Avatar upload failed:', err);
      setError(err.response?.data?.error || 'アバターのアップロードに失敗しました');
    } finally {
      setAvatarLoading(false);
    }
  };

  const handleDeleteAvatar = async () => {
    setError('');
    setAvatarLoading(true);
    try {
      const response = await settingsRepo.deleteAvatar();
      setProfileSuccess(response.message);
      const updatedProfile = await settingsRepo.getProfile();
      setProfile(updatedProfile);
      // フォームの値も更新されたプロフィールで同期
      setDisplayName(updatedProfile.display_name);
      setEmail(updatedProfile.email);
      // authStoreのユーザー情報も更新（ヘッダーのアバター表示用）
      await loadUser();
    } catch (err: any) {
      setError(err.response?.data?.error || 'アバターの削除に失敗しました');
    } finally {
      setAvatarLoading(false);
    }
  };

  const handleSendVerification = async () => {
    setError('');
    try {
      const response = await settingsRepo.sendEmailVerification();
      setProfileSuccess(response.message);
    } catch (err: any) {
      setError(err.response?.data?.error || 'メール送信に失敗しました');
    }
  };

  const handleArchiveAccount = async () => {
    if (!deletePassword) {
      setError('パスワードを入力してください');
      return;
    }

    setError('');
    try {
      await settingsRepo.archiveAccount({
        password: deletePassword,
        deletion_reason: deleteReason || undefined,
      });
      // ログアウト処理
      localStorage.removeItem('auth_token');
      navigate('/login');
    } catch (err: any) {
      setError(err.response?.data?.error || 'アカウント削除に失敗しました');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">読み込み中...</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-red-500">プロフィールを読み込めませんでした</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6 pb-20 md:pb-6">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate(-1)}
          className="mr-4 p-2 hover:bg-gray-100 rounded-full"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-bold">ユーザー設定</h1>
      </div>

      {/* タブナビゲーション */}
      <div className="bg-white rounded-xl shadow">
        <div className="flex border-b">
          <button
            onClick={() => setActiveTab('profile')}
            className={`flex-1 py-3 px-4 text-sm font-medium ${activeTab === 'profile'
              ? 'text-primary-600 border-b-2 border-primary-600'
              : 'text-gray-500 hover:text-gray-700'
              }`}
          >
            プロフィール
          </button>
          <button
            onClick={() => setActiveTab('security')}
            className={`flex-1 py-3 px-4 text-sm font-medium ${activeTab === 'security'
              ? 'text-primary-600 border-b-2 border-primary-600'
              : 'text-gray-500 hover:text-gray-700'
              }`}
          >
            セキュリティ
          </button>
          <button
            onClick={() => setActiveTab('account')}
            className={`flex-1 py-3 px-4 text-sm font-medium ${activeTab === 'account'
              ? 'text-primary-600 border-b-2 border-primary-600'
              : 'text-gray-500 hover:text-gray-700'
              }`}
          >
            アカウント
          </button>
        </div>

        <div className="p-6">
          {/* エラー・成功メッセージ */}
          {error && (
            <div className="mb-4 rounded-lg bg-red-50 p-4">
              <div className="text-sm text-red-800">{error}</div>
            </div>
          )}
          {profileSuccess && (
            <div className="mb-4 rounded-lg bg-green-50 p-4">
              <div className="text-sm text-green-800">{profileSuccess}</div>
            </div>
          )}
          {usernameSuccess && (
            <div className="mb-4 rounded-lg bg-green-50 p-4">
              <div className="text-sm text-green-800">{usernameSuccess}</div>
            </div>
          )}
          {passwordSuccess && (
            <div className="mb-4 rounded-lg bg-green-50 p-4">
              <div className="text-sm text-green-800">{passwordSuccess}</div>
            </div>
          )}

          {/* プロフィールタブ */}
          {activeTab === 'profile' && (
            <div className="space-y-6">
              {/* アバター */}
              <div>
                <h3 className="text-lg font-semibold mb-4">アバター画像</h3>
                <div className="flex items-center space-x-4">
                  <div className="w-24 h-24 rounded-full overflow-hidden bg-gray-200">
                    {avatarPreview ? (
                      <img src={avatarPreview} alt="Preview" className="w-full h-full object-cover" />
                    ) : profile.avatar_url ? (
                      <img src={profile.avatar_url} alt="Avatar" className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-2xl font-bold text-gray-500">
                        {profile.display_name[0]}
                      </div>
                    )}
                  </div>
                  <div className="flex-1 space-y-2">
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleAvatarChange}
                      className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100"
                    />
                    <div className="flex space-x-2">
                      {avatarFile && (
                        <button
                          onClick={handleUploadAvatar}
                          disabled={avatarLoading}
                          className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50"
                        >
                          {avatarLoading ? 'アップロード中...' : 'アップロード'}
                        </button>
                      )}
                      {profile.avatar_url && (
                        <button
                          onClick={handleDeleteAvatar}
                          disabled={avatarLoading}
                          className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 disabled:opacity-50"
                        >
                          削除
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              {/* プロフィール編集 */}
              <form onSubmit={handleUpdateProfile} className="space-y-4">
                <h3 className="text-lg font-semibold">基本情報</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    表示名
                  </label>
                  <input
                    type="text"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    required
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      苗字
                    </label>
                    <input
                      type="text"
                      value={lastName}
                      required
                      onChange={(e) => setLastName(e.target.value)}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      名前
                    </label>
                    <input
                      type="text"
                      value={firstName}
                      required
                      onChange={(e) => setFirstName(e.target.value)}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    メールアドレス
                  </label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    required
                  />
                  {!profile.email_verified && (
                    <div className="mt-2 flex items-center space-x-2">
                      <span className="text-sm text-yellow-600">メールアドレスが未認証です</span>
                      <button
                        type="button"
                        onClick={handleSendVerification}
                        className="text-sm text-primary-600 hover:text-primary-700"
                      >
                        認証メールを送信
                      </button>
                    </div>
                  )}
                </div>
                <button
                  type="submit"
                  disabled={profileLoading}
                  className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50"
                >
                  {profileLoading ? '更新中...' : '保存'}
                </button>
              </form>

              {/* ユーザー名変更 */}
              <form onSubmit={handleUpdateUsername} className="space-y-4 pt-6 border-t">
                <h3 className="text-lg font-semibold">ユーザー名変更</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    現在のユーザー名: <span className="font-bold">{profile.username}</span>
                  </label>
                  <input
                    type="text"
                    value={newUsername}
                    onChange={(e) => setNewUsername(e.target.value)}
                    placeholder="新しいユーザー名"
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    minLength={3}
                    maxLength={50}
                  />
                  <p className="mt-1 text-xs text-gray-500">
                    ユーザー名は3〜50文字で設定してください
                  </p>
                </div>
                <button
                  type="submit"
                  disabled={usernameLoading || !newUsername}
                  className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50"
                >
                  {usernameLoading ? '変更中...' : 'ユーザー名を変更'}
                </button>
              </form>
            </div>
          )}

          {/* セキュリティタブ */}
          {activeTab === 'security' && (
            <form onSubmit={handleChangePassword} className="space-y-4">
              <h3 className="text-lg font-semibold">パスワード変更</h3>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  現在のパスワード
                </label>
                <input
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  新しいパスワード
                </label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  required
                  minLength={8}
                />
                <p className="mt-1 text-xs text-gray-500">
                  パスワードは8文字以上にしてください
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  新しいパスワード（確認）
                </label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  required
                />
              </div>
              <button
                type="submit"
                disabled={passwordLoading}
                className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50"
              >
                {passwordLoading ? '変更中...' : 'パスワードを変更'}
              </button>
            </form>
          )}

          {/* アカウントタブ */}
          {activeTab === 'account' && (
            <div className="space-y-6">
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <h3 className="text-lg font-semibold text-yellow-800 mb-2">アカウント削除について</h3>
                <p className="text-sm text-yellow-700">
                  アカウントを削除すると、すべてのデータが削除され、復元できません。
                  この操作は取り消せませんので、十分ご注意ください。
                </p>
              </div>

              {!showDeleteConfirm ? (
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="w-full bg-red-600 text-white py-3 rounded-lg font-medium hover:bg-red-700"
                >
                  アカウントを削除
                </button>
              ) : (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      パスワードを入力して確認
                    </label>
                    <input
                      type="password"
                      value={deletePassword}
                      onChange={(e) => setDeletePassword(e.target.value)}
                      placeholder="パスワード"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      削除理由（任意）
                    </label>
                    <textarea
                      value={deleteReason}
                      onChange={(e) => setDeleteReason(e.target.value)}
                      placeholder="アカウント削除の理由をお聞かせください"
                      rows={3}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    />
                  </div>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => {
                        setShowDeleteConfirm(false);
                        setDeletePassword('');
                        setDeleteReason('');
                      }}
                      className="flex-1 bg-gray-200 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-300"
                    >
                      キャンセル
                    </button>
                    <button
                      onClick={handleArchiveAccount}
                      className="flex-1 bg-red-600 text-white py-3 rounded-lg font-medium hover:bg-red-700"
                    >
                      削除を確定
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
