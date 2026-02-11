import React, { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { User } from '@/core/domain/User';

interface LocationState {
  toUser: User;
  amount: number;
  message: string;
}

export const TransferRequestSuccessPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const state = location.state as LocationState;

  useEffect(() => {
    // 3秒後に送金リクエスト一覧に遷移
    const timer = setTimeout(() => {
      navigate('/transfer-requests');
    }, 3000);

    return () => clearTimeout(timer);
  }, [navigate]);

  if (!state || !state.toUser) {
    navigate('/dashboard');
    return null;
  }

  const { toUser, amount, message } = state;

  return (
    <div className="max-w-md mx-auto space-y-6 pb-20 md:pb-6">
      <div className="bg-white rounded-xl shadow p-8 text-center space-y-6">
        {/* 成功アイコン */}
        <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto">
          <svg className="w-10 h-10 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        </div>

        {/* タイトル */}
        <div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">送金リクエストを送信しました!</h2>
          <p className="text-gray-600">相手が承認すると送金が完了します</p>
        </div>

        {/* 送信内容 */}
        <div className="bg-gray-50 rounded-lg p-6 space-y-4">
          {/* 送信先ユーザー */}
          <div className="flex items-center justify-center space-x-3">
            <div className="w-12 h-12 bg-gradient-to-br from-primary-400 to-primary-600 rounded-full flex items-center justify-center text-white font-bold">
              {toUser.avatar_url ? (
                <img
                  src={toUser.avatar_url}
                  alt={toUser.display_name}
                  className="w-full h-full rounded-full object-cover"
                />
              ) : (
                toUser.display_name?.charAt(0).toUpperCase() || toUser.username.charAt(0).toUpperCase()
              )}
            </div>
            <div className="text-left">
              <div className="font-bold text-gray-900">{toUser.display_name || toUser.username}</div>
              <div className="text-sm text-gray-500">@{toUser.username}</div>
            </div>
          </div>

          {/* 送金額 */}
          <div className="pt-4 border-t border-gray-200">
            <div className="text-sm text-gray-600 mb-1">送金額</div>
            <div className="text-3xl font-bold text-primary-600">{amount.toLocaleString()} P</div>
          </div>

          {/* メッセージ */}
          {message && (
            <div className="pt-4 border-t border-gray-200">
              <div className="text-sm text-gray-600 mb-1">メッセージ</div>
              <div className="text-gray-900">{message}</div>
            </div>
          )}
        </div>

        {/* 自動遷移メッセージ */}
        <p className="text-sm text-gray-500">
          リクエスト一覧画面に移動します...
        </p>

        {/* ボタン */}
        <div className="space-y-3">
          <button
            onClick={() => navigate('/transfer-requests')}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors"
          >
            リクエスト一覧を見る
          </button>
          <button
            onClick={() => navigate('/dashboard')}
            className="w-full bg-white text-gray-700 py-3 rounded-lg font-medium border border-gray-300 hover:bg-gray-50 transition-colors"
          >
            ダッシュボードに戻る
          </button>
        </div>
      </div>
    </div>
  );
};
