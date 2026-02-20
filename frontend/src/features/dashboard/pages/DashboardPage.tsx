import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '@/shared/stores/authStore';
import { PointRepository } from '@/infrastructure/api/repositories/PointRepository';
import { TransferRequestRepository } from '@/infrastructure/api/repositories/TransferRequestRepository';
import { FriendshipRepository } from '@/infrastructure/api/repositories/FriendshipRepository';

const pointRepository = new PointRepository();
const transferRequestRepository = new TransferRequestRepository();
const friendshipRepository = new FriendshipRepository();

export const DashboardPage: React.FC = () => {
  const { user } = useAuthStore();
  const [balance, setBalance] = useState<number>(0);
  const [pendingRequestCount, setPendingRequestCount] = useState<number>(0);
  const [friendRequestCount, setFriendRequestCount] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [pointVisible, setPointVisible] = useState(() => {
    return localStorage.getItem('pointVisible') !== 'false';
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [balanceData, countData, friendCountData] = await Promise.all([
        pointRepository.getBalance(),
        transferRequestRepository.getPendingRequestCount(),
        friendshipRepository.getPendingRequestCount(),
      ]);
      setBalance(balanceData.balance);
      setPendingRequestCount(countData.count);
      setFriendRequestCount(friendCountData.count);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 pb-20 md:pb-6">
      {/* 残高カード */}
      <div className="bg-gradient-to-br from-primary-500 to-primary-600 rounded-2xl shadow-lg p-6 text-white">
        <div className="text-sm opacity-90 mb-2">現在の残高</div>
        {loading ? (
          <div className="h-12 flex items-center">
            <div className="animate-pulse text-2xl">読み込み中...</div>
          </div>
        ) : (
          <div className="flex items-center justify-between mb-4">
            <div className="text-4xl font-bold">
              {pointVisible ? `${balance.toLocaleString()} P` : '••••• P'}
            </div>
            <button
              onClick={() => {
                const next = !pointVisible;
                setPointVisible(next);
                localStorage.setItem('pointVisible', String(next));
              }}
              className="ml-3 p-1.5 rounded-full hover:bg-white/20 transition-colors"
              aria-label={pointVisible ? 'ポイントを隠す' : 'ポイントを表示'}
            >
              {pointVisible ? (
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              ) : (
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              )}
            </button>
          </div>
        )}
        <div className="text-sm opacity-75">
          {user?.display_name} さん
        </div>
      </div>

      {/* 送金リクエスト通知 */}
      {pendingRequestCount > 0 && (
        <Link
          to="/transfer-requests"
          className="bg-orange-50 border border-orange-200 rounded-xl p-4 flex items-center justify-between hover:bg-orange-100 transition-colors"
        >
          <div className="flex items-center">
            <div className="w-10 h-10 bg-orange-500 rounded-full flex items-center justify-center mr-3">
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            </div>
            <div>
              <div className="text-sm font-medium text-orange-900">
                新しい送金リクエスト
              </div>
              <div className="text-xs text-orange-700">
                {pendingRequestCount}件の承認待ちリクエストがあります
              </div>
            </div>
          </div>
          <svg className="w-5 h-5 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </Link>
      )}

      {/* 友達申請通知 */}
      {friendRequestCount > 0 && (
        <Link
          to="/friends"
          className="bg-blue-50 border border-blue-200 rounded-xl p-4 flex items-center justify-between hover:bg-blue-100 transition-colors"
        >
          <div className="flex items-center">
            <div className="w-10 h-10 bg-blue-500 rounded-full flex items-center justify-center mr-3">
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
              </svg>
            </div>
            <div>
              <div className="text-sm font-medium text-blue-900">
                新しい友達申請
              </div>
              <div className="text-xs text-blue-700">
                {friendRequestCount}件の友達申請が届いています
              </div>
            </div>
          </div>
          <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </Link>
      )}

      {/* アクションボタン */}
      <div className="grid grid-cols-2 gap-4">
        <Link
          to="/qr/personal"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-primary-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">マイQRコード</div>
        </Link>

        <Link
          to="/qr/scan"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">QRスキャン</div>
        </Link>

        <Link
          to="/transfer-requests"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow relative"
        >
          {pendingRequestCount > 0 && (
            <div className="absolute top-2 right-2 w-6 h-6 bg-red-500 rounded-full flex items-center justify-center">
              <span className="text-xs font-bold text-white">{pendingRequestCount > 9 ? '9+' : pendingRequestCount}</span>
            </div>
          )}
          <div className="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">送金リクエスト</div>
        </Link>

        <Link
          to="/transfer"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">友達に送金</div>
        </Link>
      </div>

      {/* クイックリンク */}
      <div className="bg-white rounded-xl shadow p-4">
        <h3 className="text-sm font-medium text-gray-900 mb-3">メニュー</h3>
        <div className="space-y-2">
          <Link
            to="/daily-bonus"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-purple-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7" />
              </svg>
              <span className="text-sm text-gray-700">デイリーボーナス</span>
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          <Link
            to="/history"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm text-gray-700">取引履歴</span>
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          <Link
            to="/friends"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              <span className="text-sm text-gray-700">友達管理</span>
              {friendRequestCount > 0 && (
                <span className="ml-2 px-2 py-0.5 text-xs font-bold text-white bg-blue-500 rounded-full">
                  {friendRequestCount > 9 ? '9+' : friendRequestCount}
                </span>
              )}
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          <Link
            to="/products"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
              <span className="text-sm text-gray-700">商品交換</span>
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          {user?.role === 'admin' && (
            <Link
              to="/admin"
              className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <div className="flex items-center">
                <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span className="text-sm text-gray-700">管理者画面</span>
              </div>
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </Link>
          )}
        </div>
      </div>
    </div>
  );
};
