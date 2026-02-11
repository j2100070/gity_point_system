import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { SettingsRepository } from '../api/SettingsRepository';

const settingsRepo = new SettingsRepository();

export const VerifyEmailPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      setError('認証トークンが見つかりません');
      setLoading(false);
      return;
    }

    verifyEmail(token);
  }, [searchParams]);

  const verifyEmail = async (token: string) => {
    try {
      await settingsRepo.verifyEmail({ token });
      setSuccess(true);
      setLoading(false);
      // 3秒後に設定ページにリダイレクト
      setTimeout(() => {
        navigate('/settings');
      }, 3000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'メール認証に失敗しました');
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mb-4"></div>
          <p className="text-gray-600">メールアドレスを認証中...</p>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-md w-full bg-white rounded-xl shadow-lg p-8 text-center">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">メール認証完了！</h2>
          <p className="text-gray-600 mb-6">
            メールアドレスが正常に認証されました。
          </p>
          <p className="text-sm text-gray-500">
            3秒後に設定ページに移動します...
          </p>
          <button
            onClick={() => navigate('/settings')}
            className="mt-4 px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
          >
            今すぐ移動
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white rounded-xl shadow-lg p-8 text-center">
        <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">認証エラー</h2>
        <p className="text-gray-600 mb-6">{error}</p>
        <div className="flex space-x-2">
          <button
            onClick={() => navigate('/settings')}
            className="flex-1 px-6 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
          >
            設定に戻る
          </button>
          <button
            onClick={() => navigate('/dashboard')}
            className="flex-1 px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
          >
            ダッシュボード
          </button>
        </div>
      </div>
    </div>
  );
};
