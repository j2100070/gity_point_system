import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { TransferRequestRepository } from '@/infrastructure/api/repositories/TransferRequestRepository';
import { PointRepository } from '@/infrastructure/api/repositories/PointRepository';

const transferRequestRepository = new TransferRequestRepository();
const pointRepository = new PointRepository();

interface ToUserInfo {
  id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  balance: number;
}

export const ConfirmTransferRequestPage: React.FC = () => {
  const [toUser, setToUser] = useState<ToUserInfo | null>(null);
  const [amount, setAmount] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [myBalance, setMyBalance] = useState(0);
  const navigate = useNavigate();
  const location = useLocation();

  // URLパラメータからユーザーIDを取得
  const searchParams = new URLSearchParams(location.search);
  const toUserId = searchParams.get('userId');

  useEffect(() => {
    if (!toUserId) {
      setError('ユーザーIDが指定されていません');
      setLoading(false);
      return;
    }

    loadMyBalance();
    setLoading(false);
  }, [toUserId]);

  const loadMyBalance = async () => {
    try {
      const data = await pointRepository.getBalance();
      setMyBalance(data.balance);
    } catch (err) {
      console.error('Failed to load balance:', err);
    }
  };

  const generateIdempotencyKey = () => {
    return `transfer-request-${Date.now()}-${Math.random().toString(36).substring(7)}`;
  };

  const handleSubmit = async () => {
    if (!amount || parseInt(amount) <= 0) {
      setError('送金額を入力してください');
      return;
    }

    if (parseInt(amount) > myBalance) {
      setError('残高が不足しています');
      return;
    }

    setError('');
    setSubmitting(true);
    try {
      const response = await transferRequestRepository.createTransferRequest({
        to_user_id: toUserId!,
        amount: parseInt(amount),
        message: message,
        idempotency_key: generateIdempotencyKey(),
      });

      // 成功時、受信者の情報を保存
      setToUser(response.to_user);

      // 成功画面に遷移
      navigate('/qr/scan/success', {
        state: {
          toUser: response.to_user,
          amount: parseInt(amount),
          message: message,
        },
      });
    } catch (err: any) {
      setError(err.response?.data?.error || '送金リクエストの作成に失敗しました');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-md mx-auto space-y-6 pb-20 md:pb-6">
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto space-y-6 pb-20 md:pb-6">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate(-1)}
          className="mr-4 p-2 hover:bg-gray-100 rounded-full"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-bold">送金リクエスト</h1>
      </div>

      <div className="bg-white rounded-xl shadow p-6 space-y-6">
        {/* 送信先ユーザー情報プレースホルダー */}
        <div className="text-center pb-6 border-b border-gray-200">
          <div className="w-20 h-20 bg-gradient-to-br from-primary-400 to-primary-600 rounded-full mx-auto mb-3 flex items-center justify-center text-white text-2xl font-bold">
            {toUser && toUser.avatar_url ? (
              <img
                src={toUser.avatar_url}
                alt={toUser.display_name}
                className="w-full h-full rounded-full object-cover"
              />
            ) : (
              <svg className="w-10 h-10" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
              </svg>
            )}
          </div>
          {toUser ? (
            <>
              <h2 className="text-lg font-bold text-gray-900">{toUser.display_name || toUser.username}</h2>
              <p className="text-sm text-gray-500">@{toUser.username}</p>
            </>
          ) : (
            <>
              <h2 className="text-lg font-bold text-gray-900">送金先ユーザー</h2>
              <p className="text-sm text-gray-500">QRコードをスキャンしました</p>
            </>
          )}
        </div>

        {/* 残高表示 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex justify-between items-center">
            <span className="text-sm text-gray-600">現在の残高</span>
            <span className="text-lg font-bold text-gray-900">{myBalance.toLocaleString()} P</span>
          </div>
        </div>

        {/* 送金額入力 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            送金額 <span className="text-red-500">*</span>
          </label>
          <div className="relative">
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="送りたいポイント数を入力"
              className="w-full px-4 py-3 text-lg border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              required
              autoFocus
            />
            <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500 font-medium">P</span>
          </div>
        </div>

        {/* メッセージ入力 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            メッセージ（任意）
          </label>
          <textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder="メモを追加できます"
            rows={3}
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>

        {error && (
          <div className="rounded-lg bg-red-50 p-4">
            <div className="text-sm text-red-800">{error}</div>
          </div>
        )}

        {/* 説明 */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start">
            <svg className="w-5 h-5 text-blue-600 mr-2 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div className="text-xs text-blue-800">
              送金リクエストを送信します。相手が「受け取る」を押すと送金が完了します。
            </div>
          </div>
        </div>

        <button
          onClick={handleSubmit}
          disabled={submitting || !amount || parseInt(amount) <= 0}
          className="w-full bg-primary-600 text-white py-4 rounded-lg font-medium text-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {submitting ? '送信中...' : '送金リクエストを送信'}
        </button>
      </div>
    </div>
  );
};
